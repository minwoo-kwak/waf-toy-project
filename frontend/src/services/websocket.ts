import { WAFLog, WAFStats } from '../types/waf';
import { WEBSOCKET_MESSAGE_TYPES, DEFAULT_VALUES } from '../constants';

interface WebSocketMessage {
  type: string;
  data: any;
  timestamp: string;
}

class WebSocketService {
  private ws: WebSocket | null = null;
  private reconnectAttempts = 0;
  private maxReconnectAttempts = DEFAULT_VALUES.MAX_RECONNECT_ATTEMPTS;
  private reconnectDelay = DEFAULT_VALUES.RECONNECT_DELAY;
  private token: string | null = null;
  private callbacks: { [key: string]: Function[] } = {};

  connect(token: string): void {
    if (this.ws && this.ws.readyState === WebSocket.OPEN) {
      return;
    }

    this.token = token;
    const WS_URL = process.env.REACT_APP_WS_URL || 'ws://localhost/api/v1/ws';
    
    try {
      this.ws = new WebSocket(`${WS_URL}?token=${token}`);
      this.setupEventHandlers();
    } catch (error) {
      console.error('WebSocket connection failed:', error);
      this.handleReconnect();
    }
  }

  private setupEventHandlers(): void {
    if (!this.ws) return;

    this.ws.onopen = () => {
      console.log('WebSocket connected');
      this.reconnectAttempts = 0;
      this.emit('connect', null);
    };

    this.ws.onclose = (event) => {
      console.log('WebSocket disconnected:', event.code, event.reason);
      this.handleReconnect();
    };

    this.ws.onerror = (error) => {
      console.error('WebSocket error:', error);
      this.handleReconnect();
    };

    this.ws.onmessage = (event) => {
      try {
        const message: WebSocketMessage = JSON.parse(event.data);
        this.handleMessage(message);
      } catch (error) {
        console.error('Failed to parse WebSocket message:', error);
      }
    };
  }

  private handleMessage(message: WebSocketMessage): void {
    switch (message.type) {
      case WEBSOCKET_MESSAGE_TYPES.WELCOME:
        this.emit('welcome', message.data);
        break;
      case WEBSOCKET_MESSAGE_TYPES.NEW_LOG:
        this.emit('new_log', message.data);
        break;
      case WEBSOCKET_MESSAGE_TYPES.STATS_UPDATE:
        this.emit('stats_update', message.data);
        break;
      case WEBSOCKET_MESSAGE_TYPES.STATS:
        this.emit('stats', message.data);
        break;
      case WEBSOCKET_MESSAGE_TYPES.LOGS:
        this.emit('logs', message.data);
        break;
      default:
        console.log('Unknown message type:', message.type);
    }
  }

  private emit(event: string, data: any): void {
    const eventCallbacks = this.callbacks[event] || [];
    eventCallbacks.forEach(callback => callback(data));
  }

  private handleReconnect(): void {
    if (this.reconnectAttempts >= this.maxReconnectAttempts) {
      console.warn('WebSocket: Max reconnection attempts reached. Please refresh the page if needed.');
      this.emit('max_reconnects_reached', null);
      return;
    }

    this.reconnectAttempts++;
    const delay = this.reconnectDelay * Math.pow(2, this.reconnectAttempts - 1);
    
    setTimeout(() => {
      console.log(`WebSocket: Attempting to reconnect (${this.reconnectAttempts}/${this.maxReconnectAttempts})`);
      if (this.token) {
        this.connect(this.token);
      }
    }, delay);
  }

  onWelcome(callback: (data: any) => void): void {
    this.addCallback('welcome', callback);
  }

  onNewLog(callback: (log: WAFLog) => void): void {
    this.addCallback('new_log', callback);
  }

  onStatsUpdate(callback: (stats: WAFStats) => void): void {
    this.addCallback('stats_update', callback);
  }

  onStats(callback: (stats: WAFStats) => void): void {
    this.addCallback('stats', callback);
  }

  onLogs(callback: (logs: WAFLog[]) => void): void {
    this.addCallback('logs', callback);
  }

  private addCallback(event: string, callback: Function): void {
    if (!this.callbacks[event]) {
      this.callbacks[event] = [];
    }
    this.callbacks[event].push(callback);
  }

  requestLogs(limit: number = DEFAULT_VALUES.LOG_LIMIT): void {
    this.sendMessage({
      type: WEBSOCKET_MESSAGE_TYPES.GET_LOGS,
      limit,
    });
  }

  requestStats(): void {
    this.sendMessage({
      type: WEBSOCKET_MESSAGE_TYPES.GET_STATS,
    });
  }

  private sendMessage(message: any): void {
    if (this.ws && this.ws.readyState === WebSocket.OPEN) {
      this.ws.send(JSON.stringify(message));
    } else {
      console.warn('WebSocket is not connected');
    }
  }

  disconnect(): void {
    if (this.ws) {
      this.ws.close();
      this.ws = null;
    }
    this.reconnectAttempts = 0;
    this.callbacks = {};
  }

  isConnected(): boolean {
    return this.ws?.readyState === WebSocket.OPEN;
  }
}

export default new WebSocketService();