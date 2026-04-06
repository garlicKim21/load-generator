import React, { useEffect, useRef } from 'react';
import { ScrollText, Info, AlertTriangle, AlertCircle, CheckCircle } from 'lucide-react';
import { StreamEvent } from '../types';

interface EventFeedProps {
  events: StreamEvent[];
}

const typeConfig: Record<string, { icon: React.FC<any>; color: string }> = {
  info: { icon: Info, color: '#60a5fa' },
  warning: { icon: AlertTriangle, color: '#fbbf24' },
  error: { icon: AlertCircle, color: '#f87171' },
  success: { icon: CheckCircle, color: '#34d399' },
};

const EventFeed: React.FC<EventFeedProps> = ({ events }) => {
  const scrollRef = useRef<HTMLDivElement>(null);

  useEffect(() => {
    if (scrollRef.current) {
      scrollRef.current.scrollTop = scrollRef.current.scrollHeight;
    }
  }, [events]);

  return (
    <div className="h-full flex flex-col bg-white/5 backdrop-blur-lg border border-white/10 rounded-2xl overflow-hidden">
      {/* Header */}
      <div className="flex items-center gap-2 px-5 py-4 border-b border-white/10">
        <ScrollText size={18} className="text-gray-400" />
        <h2 className="text-base font-semibold text-white">Event Feed</h2>
      </div>

      {/* Event list */}
      <div ref={scrollRef} className="flex-1 overflow-y-auto p-4 space-y-2 scrollbar-thin">
        {events.length === 0 && (
          <p className="text-gray-500 text-sm text-center mt-8">
            No events yet. Start a load generator to see activity.
          </p>
        )}
        {events.map((event, i) => {
          const config = typeConfig[event.type] || typeConfig.info;
          const Icon = config.icon;
          return (
            <div
              key={i}
              className="flex items-start gap-3 px-3 py-2.5 rounded-lg bg-white/5 hover:bg-white/8 transition-colors animate-slide-in"
            >
              <Icon size={15} style={{ color: config.color, marginTop: 2, flexShrink: 0 }} />
              <div className="flex-1 min-w-0">
                <p className="text-sm text-gray-200 leading-snug">{event.message}</p>
              </div>
              <span className="text-xs text-gray-500 tabular-nums whitespace-nowrap mt-0.5">
                {event.time}
              </span>
            </div>
          );
        })}
      </div>
    </div>
  );
};

export default EventFeed;
