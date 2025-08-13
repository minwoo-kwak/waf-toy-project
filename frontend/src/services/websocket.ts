import { WAFLog, WAFStats } from '../types/waf';

interface WebSocketMessage {
  type: string;
  data: any;
  timestamp: string;
}

class WebSocketService {
  private ws: WebSocket | null = null;
  private reconnectAttempts = 0;
  private maxReconnectAttempts = 5;
  private reconnectDelay = 1000;
  private token: string | null = null;
  private callbacks: { [key: string]: Function[] } = {};

  connect(token: string): void {
    if (this.ws && this.ws.readyState === WebSocket.OPEN) {
      return;
    }

    this.token = token;
    const WS_URL = process.env.REACT_APP_WS_URL || 'ws://localhost:3000/api/v1/ws';
    
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
      case 'welcome':
        this.emit('welcome', message.data);
        break;
      case 'new_log':
        this.emit('new_log', message.data);
        break;
      case 'stats_update':
        this.emit('stats_update', message.data);
        break;
      case 'stats':
        this.emit('stats', message.data);
        break;
      case 'logs':
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
      console.error('Max reconnection attempts reached');
      return;
    }

    this.reconnectAttempts++;
    const delay = this.reconnectDelay * Math.pow(2, this.reconnectAttempts - 1);
    
    setTimeout(() => {
      console.log(`Attempting to reconnect (${this.reconnectAttempts}/${this.maxReconnectAttempts})`);
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

  requestLogs(limit: number = 50): void {
    this.sendMessage({
      type: 'get_logs',
      limit,
    });
  }

  requestStats(): void {
    this.sendMessage({
      type: 'get_stats',
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