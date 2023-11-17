import base64
import io
import logging
import os
import requests
import time


from flask import Flask, request
from pydub import AudioSegment
# from faster_whisper import WhisperModel

import openai
from openai import OpenAI

logging.basicConfig(level=logging.INFO)
logger = logging.getLogger('gunicorn.error')

app = Flask(__name__)

# os.environ['OPENAI_API_KEY'] = "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"
# openaiClient = OpenAI()

openaiClient = openai.Client(api_key="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")


# env = os.environ.get('ENV', 'LOCAL')
# if env == 'PROD' or env == 'STAGING':
#     # Initialize the Whisper ASR model
#     faster_whisper_model = WhisperModel("large-v2", device="cuda", compute_type="float16")
# else:
#     # for local testing, use cpu.
#     faster_whisper_model = WhisperModel("base", device="cpu", compute_type="int8")

# for speech_to_text_whisper_gpu specifically.
class NamedBytesIO(io.BytesIO):
    def __init__(self, buffer, name=None):
        super().__init__(buffer)
        self.name = name

def pitch_shift(audio_bytes:bytes, octaves:float) -> bytes:
    # Create a stream from the bytes
    audio_stream = io.BytesIO(audio_bytes)
    # # Load the audio from the stream
    voice = AudioSegment.from_file(audio_stream, format="wav")
    new_sample_rate = int(voice.frame_rate * (2.0 ** octaves))

    # keep the same samples but tell the computer they ought to be played at the 
    # new, higher sample rate. This file sounds like a chipmunk but has a weird sample rate.
    hipitch_sound = voice._spawn(voice.raw_data, overrides={'frame_rate': new_sample_rate})

    # now we just convert it to a common sample rate (44.1k - standard audio CD) to 
    # make sure it works in regular audio players. Other than potentially losing audio quality (if
    # you set it too low - 44.1k is plenty) this should now noticeable change how the audio sounds.
    hipitch_sound = hipitch_sound.set_frame_rate(44100)
    # Export the modified voice as bytes
    # Create an AudioSegment from raw audio data
    audio_segment = AudioSegment(
        data=hipitch_sound.raw_data,
        sample_width=hipitch_sound.sample_width,
        frame_rate=hipitch_sound.frame_rate,
        channels=hipitch_sound.channels
    )

    # Export the audio as WAV format bytes
    return audio_segment.export(format="wav").read()

# def speech_to_text_whisper_gpu(wav_bytes:bytes) -> str:
#     audio_buffer = NamedBytesIO(wav_bytes, name="audio.wav")
#     segments, info = faster_whisper_model.transcribe(audio_buffer, beam_size=5)
#     transcribed_text = " ".join(segment.text for segment in segments)
#     return transcribed_text

def text_to_speech(text:str, gender:str) -> bytes:
    voiceName = "alloy"
    if gender == "Gender_Male" or gender == "male":
        voiceName = "echo"
    elif gender == "Gender_Female" or gender == "female":
        voiceName = "nova"
    response = openaiClient.audio.speech.create(
        model="tts-1",
        voice=voiceName,
        input=text,
    )
    bytesBuffer = io.BytesIO()
    for data in response.iter_bytes():
        bytesBuffer.write(data)
    return bytesBuffer.getvalue()

@app.route('/pitch_shift', methods=['POST'])
def pitch_shift_handler():
    if request.method == 'POST':
        try:
            data = request.json

            b64_encoded = data.get('b64_encoded', '')
            octaves = float(data.get('octaves', ''))

            inputBytes = base64.b64decode(b64_encoded)
            outputBytes = pitch_shift(inputBytes, octaves)

            return base64.b64encode(outputBytes)
        except Exception as e:
            return "Exception: " + str(e), 400
    else:
        return "Unsupported request method", 405
    
@app.route("/text_to_speech", methods=['POST'])
def speech_to_text_handler():
    if request.method == 'POST':
        try:
            data = request.json

            text = data.get('text', '')
            gender = data.get('gender', '')

            if len(text) > 0:
                mp3Bytes = text_to_speech(text, gender)
                b64Encoded = base64.b64encode(mp3Bytes).decode('utf-8')
                return b64Encoded

            return ""
        except Exception as e:
            return "Exception: " + str(e), 400
    else:
        return "Unsupported request method", 405


def create_thread(client):
    """
    Creates a new thread using the provided OpenAI client.

    Args:
        client: The OpenAI client instance to use for creating the thread.

    Returns:
        The created thread object.
    """
    my_thread = client.beta.threads.create()
    return my_thread



def wait_for_run_completion(client, thread_id, poll_interval=0.2, timeout=10):
    """
    Waits for the latest run in a thread to either complete or fail, with a timeout.

    Args:
        client: The OpenAI client instance.
        thread_id: The ID of the thread to monitor.
        poll_interval: Time in seconds to wait between each status check.
        timeout: Maximum time to wait for run completion in seconds.

    Returns:
        A tuple containing a boolean indicating success and an error message if any.
    """
    start_time = time.time()

    while True:
        # Check if the timeout is exceeded
        if time.time() - start_time > timeout:
            return False, "Timeout exceeded while waiting for run to complete."

        # Check the run status
        runs = client.beta.threads.runs.list(thread_id)
        latest_run = runs.data[0]
        if latest_run.status == "completed":
            return True, None  # Successful completion
        elif latest_run.status == "failed":
            return False, "Run failed."

        time.sleep(poll_interval)  # Wait before the next check


def create_message_and_run_thread(client, thread_id, assistant_id, content):
    """
    Creates a message in a thread, initiates a run, waits for completion,
    and retrieves the latest messages in the thread.

    Args:
        client: The OpenAI client instance.
        thread_id: The ID of the thread to interact with.
        assistant_id: The ID of the assistant to run.
        content: The content of the message to be sent.

    Returns:
        The latest message text or None if no message is retrieved or if an error occurs.
    """

    # Send the user's message to the thread
    start = time.time()
    client.beta.threads.messages.create(thread_id=thread_id, role="user", content=content)
    logger.info(f"Time to create message: {time.time() - start} seconds")

    # Initiate a run with the assistant
    start = time.time()
    client.beta.threads.runs.create(thread_id=thread_id, assistant_id=assistant_id)
    logger.info(f"Time to create run: {time.time() - start} seconds")

    # Wait for the run to complete or fail
    start = time.time()
    success, error = wait_for_run_completion(client, thread_id)
    if not success:
        logger.error(f"Error: {error}")
        return None
    logger.info(f"Time to wait for run completion: {time.time() - start} seconds")

    # Retrieve and process the messages from the thread
    start = time.time()
    response = client.beta.threads.messages.list(thread_id=thread_id)
    logger.info(f"Time to retrieve messages: {time.time() - start} seconds")

    # Iterate through the messages and return the latest message text
    for message in response.data:
        message_text = message.content[0].text.value
        if message_text.strip() != "":
            return message_text.strip()

    return None


@app.route('/create_thread', methods=['GET'])
def create_thread_handler():
    try:
        # Call the create_thread function to create a new thread
        new_thread = create_thread(openaiClient)
        logger.info("New thread created: " + new_thread.id)

        return {"thread_id": new_thread.id}
    except Exception as e:
        logger.error("Exception in create_thread: " + str(e))
        return "Exception: " + str(e), 500


@app.route('/create_message_and_run_thread', methods=['POST'])
def create_message_and_run_thread_handler():
    if request.method == 'POST':
        try:
            data = request.json
            thread_id = data.get('thread_id', '')
            api_key = data.get('api_key', '')
            assistant_id = data.get('assistant_id', '')
            content = data.get('content', '')

            start = time.time()
            client = openai.Client(api_key=api_key)
            logger.info(f"Time to create new openai client: {time.time() - start} seconds")

            if thread_id and assistant_id and content:
                message_response = create_message_and_run_thread(client, thread_id, assistant_id, content)
                return {"response": message_response}
            else:
                return "Missing required parameters", 400

        except Exception as e:
            logger.error("Exception in create_message_and_run_thread: " + str(e))
            return "Exception: " + str(e), 500
    else:
        return "Unsupported request method", 405


if __name__ == '__main__':
    env = os.environ.get('ENV', 'LOCAL')
    if env == 'PROD' or env == 'STAGING':
        from waitress import serve
        serve(app, host='0.0.0.0', port=int(os.environ.get('PORT', 8107)))
    else:
        app.run(host='0.0.0.0', port=int(os.environ.get('PORT', 8085)))
