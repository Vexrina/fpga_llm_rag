import type { Chat, Message } from '../types'

const CHATS_KEY = 'chats'

function loadChats(): Chat[] {
  try {
    const stored = localStorage.getItem(CHATS_KEY)
    if (!stored) return []
    const parsed = JSON.parse(stored)
    return parsed.map((chat: Chat) => ({
      ...chat,
      createdAt: new Date(chat.createdAt),
      updatedAt: new Date(chat.updatedAt),
      messages: chat.messages.map((msg: Message) => ({
        ...msg,
        timestamp: new Date(msg.timestamp),
      })),
    }))
  } catch {
    return []
  }
}

function saveChats(chats: Chat[]) {
  localStorage.setItem(CHATS_KEY, JSON.stringify(chats))
}

export function getChats(): Chat[] {
  return loadChats().sort((a, b) => 
    new Date(b.updatedAt).getTime() - new Date(a.updatedAt).getTime()
  )
}

export function getChat(id: string): Chat | undefined {
  return loadChats().find(chat => chat.id === id)
}

export function createChat(title: string = 'Новый чат'): Chat {
  const chats = loadChats()
  const newChat: Chat = {
    id: `chat-${Date.now()}`,
    title,
    messages: [],
    createdAt: new Date(),
    updatedAt: new Date(),
  }
  chats.push(newChat)
  saveChats(chats)
  return newChat
}

export function updateChat(id: string, messages: Message[]) {
  const chats = loadChats()
  const index = chats.findIndex(chat => chat.id === id)
  if (index === -1) return
  
  const title = messages.find(m => m.role === 'user')?.content.slice(0, 40) || 'Новый чат'
  chats[index] = {
    ...chats[index],
    messages,
    title,
    updatedAt: new Date(),
  }
  saveChats(chats)
}

export function deleteChat(id: string) {
  const chats = loadChats().filter(chat => chat.id !== id)
  saveChats(chats)
}

export function getOrCreateFirstChat(): Chat {
  const chats = getChats()
  if (chats.length > 0) return chats[0]
  return createChat()
}

const DEMO_CHATS_KEY = 'demo_chats_initialized'

export function initializeDemoChats() {
  if (localStorage.getItem(DEMO_CHATS_KEY)) return
  
  const demoChats: Chat[] = [
    {
      id: 'chat-1',
      title: 'Как получить доступ к кластеру ПиРСИИ?',
      messages: [
        {
          id: 'msg-1',
          role: 'user',
          content: 'Как получить доступ к кластеру ПиРСИИ?',
          timestamp: new Date(Date.now() - 60000),
        },
        {
          id: 'msg-2',
          role: 'assistant',
          content: `Документация по работе с кластером. Прежде, чем подать заявку на кластер, свяжитесь с менеджером Солодкой Марией (почта -- msolodkaya@itmo.ru, телеграмм -- @maria_sldk)

🔗 Инструкция по подключению студентов к кластеру:
Инструкция (PDF) (https://disk.yandex.ru/i/k28sZe_xws4bAg)

📁 Дополнительные руководства и материалы по работе с кластером доступны в директории public при запуске JupyterLab сервера.`,
          timestamp: new Date(Date.now() - 30000),
        },
      ],
      createdAt: new Date(Date.now() - 60000),
      updatedAt: new Date(Date.now() - 30000),
    },
    {
      id: 'chat-2',
      title: 'О стипендии',
      messages: [
        {
          id: 'msg-3',
          role: 'user',
          content: 'Как оформить стипендию?',
          timestamp: new Date(Date.now() - 3600000),
        },
        {
          id: 'msg-4',
          role: 'assistant',
          content: 'Для оформления стипендии необходимо подать заявление в деканат. Подробнее о видах стипендий и условиях их получения вы можете узнать на сайте университета.',
          timestamp: new Date(Date.now() - 3500000),
        },
      ],
      createdAt: new Date(Date.now() - 3600000),
      updatedAt: new Date(Date.now() - 3500000),
    },
    {
      id: 'chat-3',
      title: 'О поступлении',
      messages: [
        {
          id: 'msg-5',
          role: 'user',
          content: 'Какие документы нужны для поступления?',
          timestamp: new Date(Date.now() - 7200000),
        },
        {
          id: 'msg-6',
          role: 'assistant',
          content: 'Для поступления вам понадобятся: паспорт, аттестат о среднем образовании (или диплом), фотографии 3x4, медицинская справка. Точный перечень документов зависит от направления обучения.',
          timestamp: new Date(Date.now() - 7100000),
        },
      ],
      createdAt: new Date(Date.now() - 7200000),
      updatedAt: new Date(Date.now() - 7100000),
    },
    {
      id: 'chat-4',
      title: 'О ВКР',
      messages: [
        {
          id: 'msg-7',
          role: 'user',
          content: 'Как выбрать тему ВКР?',
          timestamp: new Date(Date.now() - 86400000),
        },
        {
          id: 'msg-8',
          role: 'assistant',
          content: 'Выбор темы ВКР осуществляется в начале 4 курса. Вы можете выбрать тему из предложенного списка кафедры или предложить свою тему, согласовав её с научным руководителем.',
          timestamp: new Date(Date.now() - 86300000),
        },
      ],
      createdAt: new Date(Date.now() - 86400000),
      updatedAt: new Date(Date.now() - 86300000),
    },
  ]

  saveChats(demoChats)
  localStorage.setItem(DEMO_CHATS_KEY, 'true')
}