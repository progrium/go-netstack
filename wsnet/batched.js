/**
 * BatchedWebSocket – browser WebSocket replacement that matches wsnet.BatchedConn behavior.
 *
 * - send(data): buffers and flushes when buffer reaches maxBatchBytes or after flushInterval
 *   (single frames still go out within flushInterval).
 * - read(buffer): reads received bytes into the given Uint8Array; resolves with bytes read.
 *   Incoming WebSocket messages are drained as a stream (same as Go BatchedConn Read).
 *
 * Use the same options on both ends (this client and the Go server with wsnet.BatchedConn).
 */

const DEFAULT_FLUSH_INTERVAL_MS = 2;
const DEFAULT_MAX_BATCH_BYTES = 32 * 1500;

export class BatchedWebSocket {
  /**
   * @param {string|WebSocket} urlOrWs - WebSocket URL to connect, or an existing WebSocket to wrap
   * @param {Object} [opts]
   * @param {number} [opts.flushIntervalMs] - Max ms to hold data before sending (default 2)
   * @param {number} [opts.maxBatchBytes] - Flush when buffered bytes >= this (default 32*1500)
   */
  constructor(urlOrWs, opts = {}) {
    this.flushIntervalMs = opts.flushIntervalMs ?? DEFAULT_FLUSH_INTERVAL_MS;
    this.maxBatchBytes = opts.maxBatchBytes ?? DEFAULT_MAX_BATCH_BYTES;

    this._ws = typeof urlOrWs === 'string' ? new WebSocket(urlOrWs) : urlOrWs;
    this._writeBuf = [];
    this._writeLen = 0;
    this._flushTimer = null;

    /** @type {Uint8Array[]} - received chunks not yet consumed by read() */
    this._readChunks = [];
    this._readChunkOff = 0;
    /** @type {Array<{resolve: (n: number) => void, buffer: Uint8Array}>} */
    this._readWaiters = [];

    this._ws.binaryType = 'arraybuffer';
    this._ws.onmessage = (ev) => {
      const data = ev.data;
      const view = data instanceof ArrayBuffer ? new Uint8Array(data) : new Uint8Array(data.buffer, data.byteOffset, data.byteLength);
      this._readChunks.push(view);
      this._drainReadWaiters();
    };
    this._userOnclose = null;
    this._ws.onclose = (ev) => {
      while (this._readWaiters.length > 0) {
        const { resolve } = this._readWaiters.shift();
        resolve(0);
      }
      if (this._userOnclose) this._userOnclose(ev);
    };
  }

  get readyState() {
    return this._ws.readyState;
  }

  get url() {
    return this._ws.url;
  }

  set onopen(fn) {
    this._ws.onopen = fn;
  }
  set onclose(fn) {
    this._userOnclose = fn;
  }
  set onerror(fn) {
    this._ws.onerror = fn;
  }

  /**
   * Send data; may be buffered and flushed on timer or when batch is full.
   * @param {ArrayBuffer|ArrayBufferView|Uint8Array} data
   */
  send(data) {
    const view = data instanceof ArrayBuffer ? new Uint8Array(data) : new Uint8Array(data.buffer, data.byteOffset, data.byteLength);
    this._writeBuf.push(view);
    this._writeLen += view.length;

    if (this._writeLen >= this.maxBatchBytes) {
      this._flush();
    } else {
      this._armFlushTimer();
    }
  }

  _armFlushTimer() {
    if (this._flushTimer != null) return;
    this._flushTimer = setTimeout(() => {
      this._flushTimer = null;
      this._flush();
    }, this.flushIntervalMs);
  }

  _flush() {
    if (this._flushTimer != null) {
      clearTimeout(this._flushTimer);
      this._flushTimer = null;
    }
    if (this._writeBuf.length === 0) return;

    const total = this._writeLen;
    const out = new Uint8Array(total);
    let off = 0;
    for (const chunk of this._writeBuf) {
      out.set(chunk, off);
      off += chunk.length;
    }
    this._writeBuf = [];
    this._writeLen = 0;

    if (this._ws.readyState === WebSocket.OPEN) {
      this._ws.send(out.buffer);
    }
  }

  /**
   * Read up to buffer.length bytes from the received stream; resolves with bytes read.
   * If no data is available, waits for the next WebSocket message.
   * @param {Uint8Array} buffer
   * @returns {Promise<number>}
   */
  read(buffer) {
    return new Promise((resolve) => {
      const n = this._copyOut(buffer);
      if (n > 0) {
        resolve(n);
        return;
      }
      this._readWaiters.push({ resolve, buffer });
    });
  }

  _copyOut(buffer) {
    let total = 0;
    while (total < buffer.length && this._readChunks.length > 0) {
      const chunk = this._readChunks[0];
      const rest = chunk.length - this._readChunkOff;
      const n = Math.min(rest, buffer.length - total);
      buffer.set(chunk.subarray(this._readChunkOff, this._readChunkOff + n), total);
      total += n;
      this._readChunkOff += n;
      if (this._readChunkOff >= chunk.length) {
        this._readChunks.shift();
        this._readChunkOff = 0;
      }
    }
    return total;
  }

  _drainReadWaiters() {
    while (this._readWaiters.length > 0 && this._readChunks.length > 0) {
      const { resolve, buffer } = this._readWaiters.shift();
      const n = this._copyOut(buffer);
      resolve(n);
    }
  }

  /**
   * Flush any buffered send data and close the WebSocket.
   */
  close(code, reason) {
    this._flush();
    this._ws.close(code, reason);
  }
}

export default BatchedWebSocket;
