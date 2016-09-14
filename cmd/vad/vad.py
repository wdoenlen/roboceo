import collections
import contextlib
import sys
import wave
import json

import webrtcvad


def read_wave(path):
    with contextlib.closing(wave.open(path, 'rb')) as wf:
        num_channels = wf.getnchannels()
        assert num_channels == 1
        sample_width = wf.getsampwidth()
        assert sample_width == 2
        sample_rate = wf.getframerate()
        assert sample_rate in (8000, 16000, 32000)
        pcm_data = wf.readframes(wf.getnframes())
        return pcm_data, sample_rate


def write_wave(path, audio, sample_rate):
    with contextlib.closing(wave.open(path, 'wb')) as wf:
        wf.setnchannels(1)
        wf.setsampwidth(2)
        wf.setframerate(sample_rate)
        wf.writeframes(audio)


def frame_generator(frame_duration_ms, audio, sample_rate):
    n = int(sample_rate * (frame_duration_ms / 1000.0) * 2)
    offset = 0
    while offset + n < len(audio):
        yield audio[offset:offset + n]
        offset += n


def vad_collector(sample_rate, num_padding_frames, vad, frames):
    ring_buffer = collections.deque(maxlen=num_padding_frames)
    triggered = False
    start_frame = -1
    for i, frame in enumerate(frames):
        if not triggered:
            ring_buffer.append(frame)
            num_voiced = len([f for f in ring_buffer
                              if vad.is_speech(f, sample_rate)])
            if num_voiced > 0.9 * ring_buffer.maxlen:
                triggered = True
                start_frame = i - num_padding_frames
                ring_buffer.clear()
        else:
            ring_buffer.append(frame)
            num_unvoiced = len([f for f in ring_buffer
                                if not vad.is_speech(f, sample_rate)])
            if num_unvoiced > 0.9 * ring_buffer.maxlen:
                triggered = False
                end_frame = i - num_padding_frames
                yield [start_frame, end_frame]
                ring_buffer.clear()
    if triggered:
        yield [start_frame, len(frames)]


def main(args):
    if len(args) != 2:
        sys.stderr.write(
            'Usage: example.py <aggressiveness> <path to wav file>\n')
        sys.exit(1)
    audio, sample_rate = read_wave(args[1])
    vad = webrtcvad.Vad(int(args[0]))

    frame_duration_ms = 30
    frames = frame_generator(frame_duration_ms, audio, sample_rate)
    frames = list(frames)

    results = []
    for result in vad_collector(sample_rate, 3, vad, frames):
        start_ms = result[0] * frame_duration_ms
        end_ms = result[1] * frame_duration_ms
        results.append({
            "start": float(start_ms) / 1000.0,
            "duration": float(end_ms - start_ms) / 1000.0,
        })

    print json.dumps(results, indent=2)


if __name__ == '__main__':
    main(sys.argv[1:])