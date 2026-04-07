import type { Message } from '../types'

const mockResponses: string[] = [
  'Для получения справки обратитесь в деканат или посетите раздел FAQ на сайте университета.',
  'Расписание занятий обновляется каждый семестр. Проверьте актуальную информацию в личном кабинете студента.',
  'Для оформления академического отпуска подайте заявление в деканат не позднее чем за 2 недели до начала сессии.',
  'Библиотека работает с понедельника по пятницу с 9:00 до 20:00, в субботу с 10:00 до 17:00.',
  'Для подключения к Wi-Fi используйте логин и пароль от личного кабинета студента.',
]

let messageCounter = 0

export async function sendMessage(_content: string): Promise<Message> {
  await new Promise((resolve) => setTimeout(resolve, 800 + Math.random() * 700))

  const responseText = mockResponses[messageCounter % mockResponses.length]
  messageCounter++

  return {
    id: `msg-${Date.now()}`,
    role: 'assistant',
    content: responseText,
    timestamp: new Date(),
  }
}