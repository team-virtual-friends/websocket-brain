import base64
import io
import logging
import os
import sys

from flask import Flask, request
from pydub import AudioSegment

logging.basicConfig(level=logging.INFO)
logger = logging.getLogger('gunicorn.error')

app = Flask(__name__)

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

@app.route('/pitch_shift', methods=['POST'])
def pitch_shift_handler():
    if request.method == 'POST':
        try:
            # Parse the JSON data from the request
            data = request.json

            b64_encoded = data.get('b64_encoded', '')
            octaves = float(data.get('octaves', ''))

            inputBytes = base64.b64decode(b64_encoded)
            outputBytes = pitch_shift(inputBytes, octaves)

            return base64.b64encode(outputBytes)
        except Exception as e:
            return "Invalid data format", 400
    else:
        return "Unsupported request method", 405

if __name__ == '__main__':
    env = os.environ.get('ENV', 'LOCAL')
    if env == 'PROD' or env == 'STAGING':
        from waitress import serve
        serve(app, host='0.0.0.0', port=int(os.environ.get('PORT', 8107)))
    else:
        app.run(host='0.0.0.0', port=int(os.environ.get('PORT', 8511)))
