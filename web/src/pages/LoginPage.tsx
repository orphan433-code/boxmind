import { useState, type FormEvent } from "react";
import { requestLogin, verifyLogin } from "../api/client";
import { useAuth } from "../auth/AuthContext";
import { TAGLINE } from "../brand";

export function LoginPage() {
  const { login } = useAuth();
  const [email, setEmail] = useState("");
  const [code, setCode] = useState("");
  const [step, setStep] = useState<"email" | "code">("email");
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState("");

  async function handleEmailSubmit(event: FormEvent) {
    event.preventDefault();
    setError("");
    setLoading(true);
    try {
      await requestLogin(email.trim());
      setStep("code");
    } catch (err) {
      setError(err instanceof Error ? err.message : "ошибка");
    } finally {
      setLoading(false);
    }
  }

  async function handleCodeSubmit(event: FormEvent) {
    event.preventDefault();
    setError("");
    setLoading(true);
    try {
      const result = await verifyLogin(email.trim(), code.trim());
      login(result.tokens.access_token, result.user);
    } catch (err) {
      setError(err instanceof Error ? err.message : "ошибка");
    } finally {
      setLoading(false);
    }
  }

  return (
    <div className="login-page">
      <div className="login-card">
        <header className="login-brand">
          <img className="login-logo" src="/brand-icon.png" width="64" height="64" alt="Boxmind" />
          <h1>Boxmind</h1>
          <p className="login-tagline">{TAGLINE}</p>
        </header>

        {step === "email" ? (
          <form className="login-form" onSubmit={handleEmailSubmit}>
            <div className="login-field">
              <label htmlFor="email">Email</label>
              <input
                id="email"
                type="email"
                value={email}
                onChange={(e) => setEmail(e.target.value)}
                placeholder="you@example.com"
                autoComplete="email"
                required
                autoFocus
              />
            </div>
            <button type="submit" className="login-submit" disabled={loading}>
              {loading ? "Отправляем…" : "Получить код"}
            </button>
          </form>
        ) : (
          <form className="login-form" onSubmit={handleCodeSubmit}>
            <p className="login-hint">
              Отправили код на <strong>{email}</strong>. Проверь почту.
            </p>
            <div className="login-field">
              <label htmlFor="code">Код из письма</label>
              <input
                id="code"
                className="login-code-input"
                type="text"
                inputMode="numeric"
                autoComplete="one-time-code"
                value={code}
                onChange={(e) => setCode(e.target.value)}
                placeholder="123456"
                required
                autoFocus
              />
            </div>
            <button type="submit" className="login-submit" disabled={loading}>
              {loading ? "Входим…" : "Войти"}
            </button>
            <button
              type="button"
              className="login-back link-btn"
              onClick={() => {
                setStep("email");
                setError("");
              }}
            >
              Другой email
            </button>
          </form>
        )}

        {error && (
          <p className="login-error error" role="alert">
            {error}
          </p>
        )}
      </div>
    </div>
  );
}
