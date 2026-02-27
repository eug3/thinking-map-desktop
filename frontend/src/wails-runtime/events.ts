/**
 * Wails Events Wrapper
 * Provides typed access to Wails events for real-time updates
 */

import { useEffect } from 'react';

type EventCallback = (data: any) => void;

// Event types matching the Go events
export const EventTypes = {
  NodeCreated: 'nodeCreated',
  NodeUpdated: 'nodeUpdated',
  NodeDeleted: 'nodeDeleted',
  NodeDependenciesUpdated: 'nodeDependenciesUpdated',
  MessageText: 'messageText',
  MessageConclusion: 'messageConclusion',
  MessageThought: 'messageThought',
  MessageAction: 'messageAction',
  ConclusionCompleted: 'conclusionCompleted',
  DecompositionCompleted: 'decompositionCompleted',
  Error: 'error',
} as const;

export type EventType = (typeof EventTypes)[keyof typeof EventTypes];

// Event data structures
export interface NodeEventData {
  nodeId: string;
  mapId?: string;
  status?: string;
  [key: string]: any;
}

export interface MessageEventData {
  nodeId: string;
  messageId: string;
  message: string;
  mode?: 'append' | 'replace';
  [key: string]: any;
}

// Placeholder Wails runtime functions
// These will be replaced by actual Wails bindings when the app runs
let EventsOn: (event: string, callback: (data: any) => void) => void = () => {};
let EventsOff: (event: string) => void = () => {};

// This will be initialized by Wails when the app runs
export function initWailsEvents(runtime: {
  EventsOn: (event: string, callback: (data: any) => void) => void;
  EventsOff: (event: string) => void;
}) {
  EventsOn = runtime.EventsOn;
  EventsOff = runtime.EventsOff;
}

/**
 * WailsEventManager manages event subscriptions
 */
class WailsEventManager {
  private listeners = new Map<string, Set<EventCallback>>();

  /**
   * Subscribe to an event type
   * @param eventType The event type to listen for
   * @param callback The callback function when the event is emitted
   * @returns Unsubscribe function
   */
  on(eventType: string, callback: EventCallback): () => void {
    if (!this.listeners.has(eventType)) {
      this.listeners.set(eventType, new Set());

      // Register with Wails runtime
      EventsOn(eventType, (data: any) => {
        const callbacks = this.listeners.get(eventType);
        callbacks?.forEach(cb => {
          try {
            cb(data);
          } catch (error) {
            console.error(`Error in event callback for ${eventType}:`, error);
          }
        });
      });
    }

    this.listeners.get(eventType)!.add(callback);

    // Return unsubscribe function
    return () => {
      this.off(eventType, callback);
    };
  }

  /**
   * Unsubscribe from an event type
   */
  off(eventType: string, callback: EventCallback): void {
    const callbacks = this.listeners.get(eventType);
    if (callbacks) {
      callbacks.delete(callback);
      if (callbacks.size === 0) {
        EventsOff(eventType);
        this.listeners.delete(eventType);
      }
    }
  }

  /**
   * Subscribe to an event that only fires once
   */
  once(eventType: string, callback: EventCallback): void {
    const wrapped = (data: any) => {
      callback(data);
      this.off(eventType, wrapped);
    };
    this.on(eventType, wrapped);
  }

  /**
   * Clear all listeners for an event type
   */
  clear(eventType: string): void {
    const callbacks = this.listeners.get(eventType);
    if (callbacks) {
      EventsOff(eventType);
      this.listeners.delete(eventType);
    }
  }

  /**
   * Clear all event listeners
   */
  clearAll(): void {
    this.listeners.forEach((_, eventType) => {
      EventsOff(eventType);
    });
    this.listeners.clear();
  }
}

// Export singleton instance
export const events = new WailsEventManager();

/**
 * React hook for subscribing to node events
 */
export function useNodeEvent(
  mapId: string,
  eventTypes: string[],
  callback: (eventType: string, data: any) => void,
  deps: any[] = []
) {
  useEffect(() => {
    const unsubscribes = eventTypes.map(type =>
      events.on(type, (data) => {
        if (!mapId || data.mapId === mapId || data.nodeId) {
          callback(type, data);
        }
      })
    );

    return () => {
      unsubscribes.forEach(unsub => unsub());
    };
  }, [mapId, ...deps]);
}

/**
 * React hook for subscribing to message events
 */
export function useMessageEvents(
  mapId: string,
  callback: (data: MessageEventData) => void,
  deps: any[] = []
) {
  const messageEventTypes = [
    EventTypes.MessageText,
    EventTypes.MessageConclusion,
    EventTypes.MessageThought,
  ];

  useEffect(() => {
    const unsubscribes = messageEventTypes.map(type =>
      events.on(type, (data) => {
        if (!mapId || data.mapId === mapId) {
          callback(data as MessageEventData);
        }
      })
    );

    return () => {
      unsubscribes.forEach(unsub => unsub());
    };
  }, [mapId, ...deps]);
}
