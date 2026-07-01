import { useEffect } from "react";
import { Link } from "react-router-dom";
import { useAuth } from "../auth/AuthContext";
import { TAGLINE } from "../brand";
import { usePageMeta } from "../hooks/usePageMeta";
import { usePublicPageScroll } from "../hooks/usePublicPageScroll";
import { DEFAULT_DESCRIPTION, DEFAULT_TITLE, SITE_URL } from "../seo/site";

const FEATURES = [
  {
    title: "Сохранение",
    text: "Из браузера, мессенджеров и iOS через Share.",
  },
  {
    title: "Категории",
    text: "AI предлагает раздел для каждой ссылки.",
  },
  {
    title: "Папки",
    text: "Свои коллекции для работы, учёбы и личного.",
  },
  {
    title: "Поиск",
    text: "По названию, описанию и тегам.",
  },
] as const;

const STEPS = [
  {
    title: "Добавь ссылку",
    text: "Вставь URL в вебе или отправь страницу из браузера и iPhone через Share. Всё попадает в одну библиотеку.",
  },
  {
    title: "Получи готовую карточку",
    text: "Boxmind сам подтянет название, описание и обложку страницы. Заполнять поля вручную не нужно.",
  },
  {
    title: "AI предложит категорию",
    text: "Сервис определит тип ссылки и добавит теги. При желании перенеси её в свою папку.",
  },
  {
    title: "Найди снова, когда понадобится",
    text: "Ищи по тексту, смотри разделы и открывай нужное за пару секунд, даже если прошло много времени.",
  },
] as const;

const FAQ = [
  {
    question: "Что такое Boxmind?",
    answer:
      "Boxmind - сервис для сохранения и организации ссылок. Он собирает закладки в одном месте, автоматически описывает страницы и помогает быстро возвращаться к нужному материалу. Подходит для работы, учёбы, чтения и личных подборок.",
  },
  {
    question: "Как происходит вход?",
    answer:
      "Вход по email без пароля. Указываешь почту, получаешь одноразовый код и подтверждаешь его на странице входа. После этого сессия сохраняется, и повторно вводить код нужно не каждый день.",
  },
  {
    question: "Есть ли приложение для iPhone?",
    answer:
      "Да. В iOS можно сохранять ссылки через Share и работать с закладками в мобильном приложении. Сохранённые материалы синхронизируются с веб-версией в одном аккаунте.",
  },
  {
    question: "Можно ли создавать свои папки?",
    answer:
      "Да. Папки нужны для своей структуры: работа, учёба, покупки, личные проекты. В них можно переносить уже сохранённые ссылки. AI-категории при этом остаются отдельно и не мешают ручной сортировке.",
  },
  {
    question: "Сколько это стоит?",
    answer:
      "Сейчас Boxmind бесплатен для личного использования. Можно сохранять ссылки, пользоваться поиском, папками и мобильным приложением без оплаты и привязки карты.",
  },
] as const;

function LandingJsonLd() {
  useEffect(() => {
    const payload = {
      "@context": "https://schema.org",
      "@graph": [
        {
          "@type": "WebSite",
          name: "Boxmind",
          url: SITE_URL,
          description: DEFAULT_DESCRIPTION,
          inLanguage: "ru-RU",
        },
        {
          "@type": "SoftwareApplication",
          name: "Boxmind",
          applicationCategory: "ProductivityApplication",
          operatingSystem: "Web, iOS",
          offers: {
            "@type": "Offer",
            price: "0",
            priceCurrency: "RUB",
          },
          description: DEFAULT_DESCRIPTION,
          url: SITE_URL,
        },
        {
          "@type": "FAQPage",
          mainEntity: FAQ.map((item) => ({
            "@type": "Question",
            name: item.question,
            acceptedAnswer: {
              "@type": "Answer",
              text: item.answer,
            },
          })),
        },
      ],
    };

    const script = document.createElement("script");
    script.type = "application/ld+json";
    script.id = "landing-jsonld";
    script.textContent = JSON.stringify(payload);
    document.head.appendChild(script);

    return () => {
      document.getElementById("landing-jsonld")?.remove();
    };
  }, []);

  return null;
}

export function LandingPage() {
  const { token } = useAuth();
  const ctaHref = token ? "/app" : "/login";

  usePublicPageScroll();
  usePageMeta({
    title: DEFAULT_TITLE,
    description: DEFAULT_DESCRIPTION,
    path: "/",
  });

  return (
    <div className="landing">
      <LandingJsonLd />

      <header className="landing-top">
        <div className="landing-shell landing-top-inner">
          <Link className="landing-logo" to="/" aria-label="Boxmind, на главную">
            <img src="/brand-icon.png" width="32" height="32" alt="" />
            <span>Boxmind</span>
          </Link>
          <nav aria-label="Навигация">
            <Link
              className={token ? "landing-top-cta" : "landing-top-cta landing-top-cta-ghost"}
              to={ctaHref}
            >
              {token ? "В приложение" : "Войти"}
            </Link>
          </nav>
        </div>
      </header>

      <main className="landing-main" id="main-content">
        <section className="landing-hero landing-shell" aria-labelledby="landing-hero-title">
          <div className="landing-hero-copy">
            <p className="landing-kicker">Умные закладки</p>
            <h1 id="landing-hero-title">Все ссылки в одном месте</h1>
            <p className="landing-lead">
              {TAGLINE}. Сохраняй из браузера и телефона, находи нужное без хаоса в заметках и
              вкладках.
            </p>
            {!token && (
              <Link className="landing-btn" to="/login">
                Начать
              </Link>
            )}
          </div>

          <div className="landing-preview" aria-hidden>
            <div className="landing-preview-card landing-preview-card-1">
              <span className="landing-preview-tag">Полезное</span>
              <strong>Документация React</strong>
              <p>Официальный справочник по хукам и API.</p>
            </div>
            <div className="landing-preview-card landing-preview-card-2">
              <span className="landing-preview-tag">Работа</span>
              <strong>Notion для команды</strong>
              <p>База знаний и задачи в одном пространстве.</p>
            </div>
            <div className="landing-preview-card landing-preview-card-3">
              <span className="landing-preview-tag">Читать</span>
              <strong>Статья про продуктивность</strong>
              <p>Сохранено на потом с тегами и поиском.</p>
            </div>
          </div>
        </section>

        <section
          className="landing-section landing-section-surface landing-shell"
          aria-labelledby="landing-features-title"
        >
          <h2 id="landing-features-title">Возможности</h2>
          <ul className="landing-features">
            {FEATURES.map((feature) => (
              <li key={feature.title}>
                <h3>{feature.title}</h3>
                <p>{feature.text}</p>
              </li>
            ))}
          </ul>
        </section>

        <section
          className="landing-section landing-section-surface landing-shell"
          aria-labelledby="landing-steps-title"
        >
          <div className="landing-section-intro">
            <h2 id="landing-steps-title">Как это работает</h2>
            <p>От сохранения ссылки до удобного поиска. Без ручной сортировки каждый раз.</p>
          </div>
          <ol className="landing-flow">
            {STEPS.map((step, index) => (
              <li key={step.title} className="landing-flow-item">
                <span className="landing-flow-num">{index + 1}</span>
                <div className="landing-flow-copy">
                  <h3>{step.title}</h3>
                  <p>{step.text}</p>
                </div>
              </li>
            ))}
          </ol>
        </section>

        <section
          className="landing-section landing-section-faq landing-shell"
          aria-labelledby="landing-faq-title"
        >
          <h2 id="landing-faq-title">Вопросы</h2>
          <div className="landing-faq">
            {FAQ.map((item) => (
              <details key={item.question} className="landing-faq-item">
                <summary>{item.question}</summary>
                <p>{item.answer}</p>
              </details>
            ))}
          </div>
        </section>
      </main>

      <footer className="landing-footer landing-shell">
        <p>© {new Date().getFullYear()} Boxmind</p>
      </footer>
    </div>
  );
}
