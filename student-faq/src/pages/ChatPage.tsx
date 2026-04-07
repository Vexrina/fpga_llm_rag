import { useState, useEffect, useRef, type FormEvent } from 'react'
import type { Message, Chat } from '../types'
import { sendMessage } from '../mocks/chat'
import { getChats, getOrCreateFirstChat, createChat, updateChat, deleteChat, initializeDemoChats } from '../mocks/chats'
import Sidebar from '../components/Sidebar'

export default function ChatPage() {
  const [chats, setChats] = useState<Chat[]>([])
  const [activeChat, setActiveChat] = useState<Chat | null>(null)
  const [input, setInput] = useState('')
  const [isLoading, setIsLoading] = useState(false)
  const messagesEndRef = useRef<HTMLDivElement>(null)
  const prevMessagesLengthRef = useRef<number>(0)

  useEffect(() => {
    initializeDemoChats()
    const loadedChats = getChats()
    setChats(loadedChats)
    if (loadedChats.length === 0) {
      const newChat = getOrCreateFirstChat()
      setActiveChat(newChat)
      setChats([newChat])
    } else {
      setActiveChat(loadedChats[0])
    }
  }, [])

  useEffect(() => {
    const messages = activeChat?.messages || []
    if (messages.length > prevMessagesLengthRef.current) {
      messagesEndRef.current?.scrollIntoView({ behavior: 'smooth' })
    }
    prevMessagesLengthRef.current = messages.length
  }, [activeChat?.messages])

  const handleSelectChat = (id: string) => {
    const chat = chats.find(c => c.id === id)
    if (chat) setActiveChat(chat)
  }

  const handleNewChat = () => {
    const newChat = createChat()
    setChats(prev => [newChat, ...prev])
    setActiveChat(newChat)
  }

  const handleDeleteChat = (id: string) => {
    deleteChat(id)
    const newChats = chats.filter(c => c.id !== id)
    setChats(newChats)
    if (activeChat?.id === id) {
      setActiveChat(newChats[0] || null)
    }
  }

  const handleSubmit = async (e: FormEvent) => {
    e.preventDefault()
    if (!input.trim() || isLoading || !activeChat) return

    const userMessage: Message = {
      id: `msg-${Date.now()}`,
      role: 'user',
      content: input.trim(),
      timestamp: new Date(),
    }

    const updatedMessages = [...activeChat.messages, userMessage]
    setActiveChat({ ...activeChat, messages: updatedMessages })
    updateChat(activeChat.id, updatedMessages)
    setChats(getChats())
    
    setInput('')
    setIsLoading(true)

    try {
      const response = await sendMessage(userMessage.content)
      const finalMessages = [...updatedMessages, response]
      setActiveChat({ ...activeChat, messages: finalMessages })
      updateChat(activeChat.id, finalMessages)
      setChats(getChats())
    } catch {
      const errorMessage: Message = {
        id: `msg-err-${Date.now()}`,
        role: 'assistant',
        content: 'Произошла ошибка. Попробуйте снова.',
        timestamp: new Date(),
      }
      const finalMessages = [...updatedMessages, errorMessage]
      setActiveChat({ ...activeChat, messages: finalMessages })
      updateChat(activeChat.id, finalMessages)
      setChats(getChats())
    } finally {
      setIsLoading(false)
    }
  }

  const messages = activeChat?.messages || []

  return (
    <div className="flex h-[calc(100vh-4rem)]">
      <Sidebar
        chats={chats}
        activeChatId={activeChat?.id || ''}
        onSelectChat={handleSelectChat}
        onNewChat={handleNewChat}
        onDeleteChat={handleDeleteChat}
      />
      
      <div className="flex-1 flex flex-col max-w-3xl mx-auto px-4 py-6 w-full">
        <div className="flex-1 overflow-y-auto space-y-4 mb-4">
          {messages.length === 0 ? (
            <div className="text-center text-gray-400 mt-16">
              <p className="text-lg">Задайте вопрос</p>
              <p className="text-sm mt-1">Мы постараемся помочь</p>
            </div>
          ) : (
            messages.map((msg) => (
              <div
                key={msg.id}
                className={`flex ${msg.role === 'user' ? 'justify-end' : 'justify-start'}`}
              >
                <div
                  className={`max-w-[80%] rounded-2xl px-4 py-2.5 ${
                    msg.role === 'user'
                      ? 'bg-indigo-600 text-white rounded-br-md'
                      : 'bg-white border border-gray-200 text-gray-800 rounded-bl-md shadow-sm'
                  }`}
                >
                  <p className="text-sm whitespace-pre-wrap">{msg.content}</p>
                  <p
                    className={`text-xs mt-1 ${
                      msg.role === 'user' ? 'text-indigo-200' : 'text-gray-400'
                    }`}
                  >
                    {new Date(msg.timestamp).toLocaleTimeString([], {
                      hour: '2-digit',
                      minute: '2-digit',
                    })}
                  </p>
                </div>
              </div>
            ))
          )}

          {isLoading && (
            <div className="flex justify-start">
              <div className="bg-white border border-gray-200 rounded-2xl rounded-bl-md px-4 py-3 shadow-sm">
                <div className="flex gap-1.5">
                  <span className="w-2 h-2 bg-gray-400 rounded-full animate-bounce" style={{ animationDelay: '0ms' }} />
                  <span className="w-2 h-2 bg-gray-400 rounded-full animate-bounce" style={{ animationDelay: '150ms' }} />
                  <span className="w-2 h-2 bg-gray-400 rounded-full animate-bounce" style={{ animationDelay: '300ms' }} />
                </div>
              </div>
            </div>
          )}

          <div ref={messagesEndRef} />
        </div>

        <form onSubmit={handleSubmit} className="flex gap-2">
          <input
            type="text"
            value={input}
            onChange={(e) => setInput(e.target.value)}
            placeholder="Введите вопрос..."
            disabled={isLoading}
            className="flex-1 rounded-xl border border-gray-300 px-4 py-2.5 text-sm focus:outline-none focus:ring-2 focus:ring-indigo-500 focus:border-transparent disabled:opacity-50 disabled:cursor-not-allowed"
          />
          <button
            type="submit"
            disabled={isLoading || !input.trim()}
            className="bg-indigo-600 text-white px-5 py-2.5 rounded-xl text-sm font-medium hover:bg-indigo-700 transition-colors disabled:opacity-50 disabled:cursor-not-allowed"
          >
            Отправить
          </button>
        </form>
      </div>
    </div>
  )
}