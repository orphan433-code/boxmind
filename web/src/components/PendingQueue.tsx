import type { PendingBookmark } from '../types'

type Props = {
  items: PendingBookmark[]
  onDismiss: (id: string) => void
}

export function PendingQueue({ items, onDismiss }: Props) {
  if (items.length === 0) return null

  return (
    <div className="pending-queue">
      {items.map((item) => (
        <div key={item.id} className={item.status === 'error' ? 'pending-item error' : 'pending-item'}>
          <div className="pending-body">
            {item.status === 'pending' ? (
              <>
                <span className="pending-spinner" aria-hidden />
                <span>AI обрабатывает…</span>
              </>
            ) : (
              <span>Не сохранилось: {item.error}</span>
            )}
            <span className="pending-url">{item.url}</span>
          </div>
          {item.status === 'error' && (
            <button type="button" className="ghost-btn small-btn" onClick={() => onDismiss(item.id)}>
              Скрыть
            </button>
          )}
        </div>
      ))}
    </div>
  )
}
