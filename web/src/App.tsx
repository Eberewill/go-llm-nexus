import { useState, useEffect } from 'react';
import { Send, Bot, User, Loader2, Sparkles, LogIn } from 'lucide-react';

interface UsageMetrics {
  promptTokens: number;
  completionTokens: number;
  totalTokens: number;
  costUsd?: number;
}

interface Message {
  role: 'user' | 'assistant';
  content: string;
  provider?: string;
  usage?: UsageMetrics;
}

interface ApiUsagePayload {
  prompt_tokens: number;
  completion_tokens: number;
  total_tokens: number;
  cost_usd?: number | null;
}

interface GenerateResponse {
  content: string;
  provider_used: string;
  processing_time_ms: number;
  usage?: ApiUsagePayload;
}

function App() {
  const [input, setInput] = useState('');
  const [messages, setMessages] = useState<Message[]>([]);
  const [isLoading, setIsLoading] = useState(false);
  const [provider, setProvider] = useState('openai');
  const [serverStatus, setServerStatus] = useState<'healthy' | 'unhealthy' | 'checking'>('checking');
  const [userId, setUserId] = useState<string | null>(null);
  const [userName, setUserName] = useState('');
  const [nameInput, setNameInput] = useState('');
  const [isRegistering, setIsRegistering] = useState(false);
  const [showRegistration, setShowRegistration] = useState(false);
  const [registrationError, setRegistrationError] = useState<string | null>(null);

  const apiBase = (import.meta.env.VITE_API_BASE || '').replace(/\/$/, '');
  const apiUrl = (path: string) => `${apiBase}${path}`;

  // Health check
  useEffect(() => {
    const endpoint = apiUrl('/api/health');
    const checkHealth = async () => {
      try {
        const response = await fetch(endpoint);
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
  }, [apiBase]);

  // Load persisted user identity
  useEffect(() => {
    const storedId = localStorage.getItem('llm-nexus-user-id');
    const storedName = localStorage.getItem('llm-nexus-user-name');
    if (storedId && storedName) {
      setUserId(storedId);
      setUserName(storedName);
      setShowRegistration(false);
    } else {
      setShowRegistration(true);
    }
  }, []);

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    if (!input.trim()) return;
    if (!userId) {
      setShowRegistration(true);
      return;
    }

    const userMsg: Message = { role: 'user', content: input };
    setMessages(prev => [...prev, userMsg]);
    setInput('');
    setIsLoading(true);

    try {
      const response = await fetch(apiUrl('/api/generate'), {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({
          user_id: userId,
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

      const data: GenerateResponse = await response.json();
      const usage: UsageMetrics | undefined = data.usage
        ? {
            promptTokens: data.usage.prompt_tokens,
            completionTokens: data.usage.completion_tokens,
            totalTokens: data.usage.total_tokens,
            costUsd: data.usage.cost_usd ?? undefined,
          }
        : undefined;

      const botMsg: Message = {
        role: 'assistant',
        content: data.content,
        provider: data.provider_used,
        usage,
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

  const registerUser = async (e: React.FormEvent) => {
    e.preventDefault();
    if (!nameInput.trim()) {
      setRegistrationError('Please enter your name.');
      return;
    }

    try {
      setIsRegistering(true);
      setRegistrationError(null);
      const response = await fetch(apiUrl('/api/users'), {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({ name: nameInput.trim() })
      });

      if (!response.ok) {
        const errorText = await response.text();
        throw new Error(errorText || 'Failed to register');
      }

      const data = await response.json();
      setUserId(data.id);
      setUserName(data.name);
      localStorage.setItem('llm-nexus-user-id', data.id);
      localStorage.setItem('llm-nexus-user-name', data.name);
      setNameInput('');
      setShowRegistration(false);
    } catch (error) {
      const message = error instanceof Error ? error.message : 'Registration failed';
      setRegistrationError(message);
    } finally {
      setIsRegistering(false);
    }
  };

  const clearIdentity = () => {
    localStorage.removeItem('llm-nexus-user-id');
    localStorage.removeItem('llm-nexus-user-name');
    setUserId(null);
    setUserName('');
    setShowRegistration(true);
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
            <div className="flex items-center gap-3">
              {userId ? (
                <div className="text-right">
                  <p className="text-xs uppercase text-stone-400">Signed in as</p>
                  <div className="flex items-center gap-2">
                    <span className="text-sm font-semibold text-stone-700">{userName}</span>
                    <button
                      type="button"
                      onClick={clearIdentity}
                      className="text-xs text-amber-700 hover:text-amber-900 font-medium"
                    >
                      Switch
                    </button>
                  </div>
                </div>
              ) : (
                <button
                  type="button"
                  onClick={() => setShowRegistration(true)}
                  className="flex items-center gap-2 bg-white/80 border border-amber-300 rounded-xl px-4 py-2 text-sm font-medium text-amber-700 hover:bg-white"
                >
                  <LogIn size={16} /> Set Name
                </button>
              )}

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
                  <div className={`w-full max-w-full sm:max-w-[75%] rounded-2xl px-5 py-3 shadow-md transition-all hover:shadow-lg ${msg.role === 'user'
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
                    <p className="whitespace-pre-wrap leading-relaxed break-words">{msg.content}</p>
                    {msg.usage && (
                      <div className={`mt-3 text-[11px] uppercase tracking-wide flex flex-wrap gap-3 ${msg.role === 'user' ? 'text-amber-100/80' : 'text-amber-700/80'}`}>
                        <span>
                          Tokens: {msg.usage.promptTokens} in / {msg.usage.completionTokens} out (total {msg.usage.totalTokens})
                        </span>
                        {typeof msg.usage.costUsd === 'number' && msg.usage.costUsd > 0 && (
                          <span>
                            Cost: ${msg.usage.costUsd.toFixed(5)}
                          </span>
                        )}
                      </div>
                    )}
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
                  placeholder={userId ? 'Type your message...' : 'Enter your name to start chatting'}
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

      {showRegistration && (
        <div className="absolute inset-0 bg-black/40 backdrop-blur-sm flex items-center justify-center z-30">
          <form onSubmit={registerUser} className="bg-white rounded-2xl w-full max-w-md p-8 shadow-2xl border border-amber-200">
            <h2 className="text-2xl font-bold text-stone-800 mb-2">Welcome to LLM Nexus</h2>
            <p className="text-sm text-stone-500 mb-6">
              Add your name so we can keep a secure audit log of your prompts.
            </p>
            <input
              type="text"
              value={nameInput}
              onChange={(e) => setNameInput(e.target.value)}
              placeholder="Your name"
              className="w-full border-2 border-amber-200/60 rounded-xl px-4 py-3 mb-3 focus:outline-none focus:ring-2 focus:ring-amber-400"
            />
            {registrationError && (
              <p className="text-sm text-red-600 mb-3">{registrationError}</p>
            )}
            <div className="flex gap-3">
              <button
                type="submit"
                disabled={isRegistering}
                className="flex-1 bg-gradient-to-br from-amber-600 to-amber-700 text-white rounded-xl py-3 font-semibold shadow hover:shadow-lg disabled:opacity-50"
              >
                {isRegistering ? 'Registering...' : 'Save & Continue'}
              </button>
              {userId && (
                <button
                  type="button"
                  onClick={() => setShowRegistration(false)}
                  className="flex-1 border-2 border-stone-200 rounded-xl py-3 font-semibold text-stone-600"
                >
                  Cancel
                </button>
              )}
            </div>
          </form>
        </div>
      )}

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
