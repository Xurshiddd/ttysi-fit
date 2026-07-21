export interface ToastItem {
  id: number
  title: string
  type: 'success' | 'error' | 'info'
}

export const useToasts = () => useState<ToastItem[]>('toasts', () => [])

export const useToast = () => {
  const toasts = useToasts()

  function add(title: string, type: ToastItem['type'] = 'info') {
    const id = Date.now() + Math.random()
    toasts.value = [...toasts.value, { id, title, type }]
    setTimeout(() => {
      toasts.value = toasts.value.filter((t) => t.id !== id)
    }, 3500)
  }

  return { add }
}
