/**
 * AudioWorklet processor: converts Float32 PCM to Int16 and posts
 * the raw ArrayBuffer to the main thread for WebSocket transmission.
 *
 * Registered as "pcm-processor" via AudioWorkletNode.
 */
class PcmProcessor extends AudioWorkletProcessor {
  process(inputs) {
    const input = inputs[0];
    if (!input || !input[0]) return true;

    const float32 = input[0]; // mono channel
    const int16 = new Int16Array(float32.length);

    for (let i = 0; i < float32.length; i++) {
      // Clamp to [-1, 1] then scale to Int16 range
      const s = Math.max(-1, Math.min(1, float32[i]));
      int16[i] = s < 0 ? s * 0x8000 : s * 0x7fff;
    }

    // Transfer the underlying buffer (zero-copy)
    this.port.postMessage(int16.buffer, [int16.buffer]);
    return true;
  }
}

registerProcessor('pcm-processor', PcmProcessor);
