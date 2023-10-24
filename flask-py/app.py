import base64
import io
import logging
import os
import requests

from flask import Flask, request
from pydub import AudioSegment
from faster_whisper import WhisperModel

logging.basicConfig(level=logging.INFO)
logger = logging.getLogger('gunicorn.error')

app = Flask(__name__)

env = os.environ.get('ENV', 'LOCAL')
if env == 'PROD' or env == 'STAGING':
    # Initialize the Whisper ASR model
    faster_whisper_model = WhisperModel("base", device="cuda", compute_type="float16")
else:
    # for local testing, use cpu.
    faster_whisper_model = WhisperModel("base", device="cpu", compute_type="int8")

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

def speech_to_text_whisper_gpu(wav_bytes:bytes) -> str:
    audio_buffer = NamedBytesIO(wav_bytes, name="audio.wav")
    segments, info = faster_whisper_model.transcribe(audio_buffer, beam_size=5)
    transcribed_text = " ".join(segment.text for segment in segments)
    return transcribed_text

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
    
@app.route("/speech_to_text", methods=['POST'])
def speech_to_text_handler():
    if request.method == 'POST':
        try:
            data = request.json

            b64_encoded = data.get('b64_encoded', '')
            inputBytes = base64.b64decode(b64_encoded)

            return speech_to_text_whisper_gpu(inputBytes)
        except Exception as e:
            return "Exception: " + str(e), 400
    else:
        return "Unsupported request method", 405

if __name__ == '__main__':
    env = os.environ.get('ENV', 'LOCAL')
    if env == 'PROD' or env == 'STAGING':
        from waitress import serve
        serve(app, host='0.0.0.0', port=int(os.environ.get('PORT', 8107)))
    else:
        app.run(host='0.0.0.0', port=int(os.environ.get('PORT', 8511)))
