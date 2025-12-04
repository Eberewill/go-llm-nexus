import { useState, useEffect } from 'react';
import { Send, Bot, User, Loader2, Sparkles } from 'lucide-react';

interface Message {
  role: 'user' | 'assistant';
  content: string;
  provider?: string;
}

function App() {
  const [input, setInput] = useState('');
  const [messages, setMessages] = useState<Message[]>([]);
  const [isLoading, setIsLoading] = useState(false);
  const [provider, setProvider] = useState('openai');
  const [serverStatus, setServerStatus] = useState<'healthy' | 'unhealthy' | 'checking'>('checking');

  // Health check
  useEffect(() => {
    const checkHealth = async () => {
      try {
        const response = await fetch('http://localhost:8080/api/health');
        if (response.ok) {
          setServerStatus('healthy');
        } else {
          setServerStatus('unhealthy');
        }
      } catch (error) {
        setServerStatus('unhealthy');
      }
    };

    checkHealth();
    const interval = setInterval(checkHealth, 10000);
    return () => clearInterval(interval);
  }, []);

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    if (!input.trim()) return;

    const userMsg: Message = { role: 'user', content: input };
    setMessages(prev => [...prev, userMsg]);
    setInput('');
    setIsLoading(true);

    try {
      const response = await fetch('http://localhost:8080/api/generate', {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({
          prompt: input,
          provider: provider,
          temperature: 0.7,
          max_tokens: 100
        }),
      });

      if (!response.ok) {
        const errorText = await response.text();
        throw new Error(`Server error: ${errorText}`);
      }

      const data = await response.json();

      const botMsg: Message = {
        role: 'assistant',
        content: data.content,
        provider: data.provider_used
      };
      setMessages(prev => [...prev, botMsg]);
    } catch (error) {
      console.error('Error:', error);
      const errorMessage = error instanceof Error ? error.message : 'Error connecting to server.';
      setMessages(prev => [...prev, { role: 'assistant', content: errorMessage }]);
    } finally {
      setIsLoading(false);
    }
  };

  return (
    <div className="h-screen bg-gradient-to-br from-amber-50 via-stone-100 to-neutral-200 flex flex-col relative overflow-hidden">
      {/* Animated background elements */}
      <div className="absolute inset-0 overflow-hidden pointer-events-none">
        <div className="absolute top-20 left-10 w-72 h-72 bg-amber-200/30 rounded-full blur-3xl animate-pulse" />
        <div className="absolute bottom-20 right-10 w-96 h-96 bg-yellow-100/20 rounded-full blur-3xl animate-pulse delay-1000" />
      </div>

      {/* Sticky Header */}
      <header className="sticky top-0 z-20 backdrop-blur-md bg-white/60 border-b border-amber-200/50 shadow-lg">
        <div className="max-w-4xl mx-auto px-6 py-4">
          <div className="flex justify-between items-center">
            <div className="flex items-center gap-4">
              <div className="relative">
                <Sparkles className="w-8 h-8 text-amber-600 animate-pulse" />
                <div className="absolute inset-0 bg-amber-400/20 blur-xl rounded-full" />
              </div>
              <div>
                <h1 className="text-3xl font-bold bg-gradient-to-r from-amber-700 via-yellow-600 to-amber-800 bg-clip-text text-transparent">
                  LLM Nexus
                </h1>
                <p className="text-xs text-stone-600 mt-1">Production-Grade AI Gateway</p>
              </div>
              <div className="flex items-center gap-2 ml-4">
                <div className={`w-2.5 h-2.5 rounded-full transition-all duration-500 ${serverStatus === 'healthy' ? 'bg-emerald-500 shadow-lg shadow-emerald-500/50' :
                  serverStatus === 'unhealthy' ? 'bg-red-500 shadow-lg shadow-red-500/50' :
                    'bg-amber-400 animate-pulse shadow-lg shadow-amber-400/50'
                  }`} />
                <span className="text-xs font-medium text-stone-700">
                  {serverStatus === 'healthy' ? 'Online' :
                    serverStatus === 'unhealthy' ? 'Offline' :
                      'Connecting...'}
                </span>
              </div>
            </div>
            <select
              value={provider}
              onChange={(e) => setProvider(e.target.value)}
              className="bg-gradient-to-br from-amber-50 to-yellow-50 border-2 border-amber-300/50 rounded-xl px-4 py-2 text-sm font-medium text-stone-800 focus:outline-none focus:ring-2 focus:ring-amber-400 focus:border-transparent transition-all hover:shadow-md cursor-pointer"
            >
              <option value="openai">OpenAI GPT</option>
              <option value="gemini">Google Gemini</option>
            </select>
          </div>
        </div>
      </header>

      {/* Main Content Area */}
      <div className="flex-1 flex items-center justify-center p-6 relative z-10 overflow-hidden">
        <div className="w-full max-w-4xl h-full flex flex-col">
          {/* Chat Container */}
          <div className="flex-1 backdrop-blur-md bg-white/60 rounded-3xl shadow-2xl overflow-hidden border border-amber-200/50 flex flex-col">
            {/* Messages */}
            <div className="flex-1 overflow-y-auto p-6 space-y-4 scroll-smooth">
              {messages.length === 0 && (
                <div className="h-full flex flex-col items-center justify-center text-stone-500">
                  <div className="relative mb-4">
                    <Bot size={64} className="text-amber-600/40" />
                    <div className="absolute inset-0 bg-amber-400/10 blur-2xl rounded-full animate-pulse" />
                  </div>
                  <p className="text-lg font-medium">Welcome to LLM Nexus</p>
                  <p className="text-sm text-stone-400 mt-2">Start a conversation with AI...</p>
                </div>
              )}

              {messages.map((msg, idx) => (
                <div
                  key={idx}
                  className={`flex ${msg.role === 'user' ? 'justify-end' : 'justify-start'} animate-fadeIn`}
                  style={{ animationDelay: `${idx * 50}ms` }}
                >
                  <div className={`max-w-[75%] rounded-2xl px-5 py-3 shadow-md transition-all hover:shadow-lg ${msg.role === 'user'
                    ? 'bg-gradient-to-br from-amber-600 to-amber-700 text-white rounded-br-sm'
                    : 'bg-gradient-to-br from-stone-50 to-amber-50 text-stone-900 border border-amber-200/50 rounded-bl-sm'
                    }`}>
                    <div className="flex items-center gap-2 mb-2 opacity-80 text-xs font-medium">
                      {msg.role === 'user' ? (
                        <User size={14} className="text-amber-100" />
                      ) : (
                        <Bot size={14} className="text-amber-700" />
                      )}
                      <span className={msg.role === 'user' ? 'text-amber-100' : 'text-amber-700'}>
                        {msg.role === 'user' ? 'You' : msg.provider || 'Assistant'}
                      </span>
                    </div>
                    <p className="whitespace-pre-wrap leading-relaxed">{msg.content}</p>
                  </div>
                </div>
              ))}

              {isLoading && (
                <div className="flex justify-start animate-fadeIn">
                  <div className="bg-gradient-to-br from-stone-50 to-amber-50 border border-amber-200/50 rounded-2xl rounded-bl-sm px-5 py-3 flex items-center gap-3 shadow-md">
                    <Loader2 size={18} className="animate-spin text-amber-600" />
                    <span className="text-sm text-stone-600 font-medium">Thinking...</span>
                  </div>
                </div>
              )}
            </div>

            {/* Input Area */}
            <form onSubmit={handleSubmit} className="p-6 border-t border-amber-200/50 bg-gradient-to-br from-amber-50/50 to-stone-50/50 backdrop-blur-sm">
              <div className="flex gap-3">
                <input
                  type="text"
                  value={input}
                  onChange={(e) => setInput(e.target.value)}
                  placeholder="Type your message..."
                  className="flex-1 bg-white/80 border-2 border-amber-200/50 rounded-xl px-5 py-3 focus:outline-none focus:ring-2 focus:ring-amber-400 focus:border-transparent transition-all placeholder:text-stone-400 text-stone-900 shadow-sm hover:shadow-md"
                />
                <button
                  type="submit"
                  disabled={isLoading || !input.trim()}
                  className="bg-gradient-to-br from-amber-600 to-amber-700 hover:from-amber-700 hover:to-amber-800 disabled:opacity-50 disabled:cursor-not-allowed text-white rounded-xl px-6 py-3 transition-all shadow-md hover:shadow-lg active:scale-95 font-medium flex items-center gap-2"
                >
                  <Send size={20} />
                </button>
              </div>
            </form>
          </div>
        </div>
      </div>

      <style>{`
        @keyframes fadeIn {
          from {
            opacity: 0;
            transform: translateY(10px);
          }
          to {
            opacity: 1;
            transform: translateY(0);
          }
        }
        .animate-fadeIn {
          animation: fadeIn 0.3s ease-out forwards;
        }
      `}</style>
    </div>
  );
}

export default App;
