import { useState, type FormEvent } from 'react'

type Props = {
  onSubmit: (url: string) => void
}

export function AddBookmarkForm({ onSubmit }: Props) {
  const [url, setUrl] = useState('')

  function handleSubmit(event: FormEvent) {
    event.preventDefault()
    const value = url.trim()
    if (!value) return

    onSubmit(value)
    setUrl('')
  }

  return (
    <form className="add-form" onSubmit={handleSubmit}>
      <span className="add-form-icon" aria-hidden>
        +
      </span>
      <input
        type="url"
        value={url}
        onChange={(e) => setUrl(e.target.value)}
        placeholder="Вставь ссылку…"
        aria-label="URL закладки"
        autoComplete="off"
        required
      />
      <button type="submit" className="add-form-btn">
        Сохранить
      </button>
    </form>
  )
}
