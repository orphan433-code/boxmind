type Props = {
  category: string;
};

export function CategoryIcon({ category }: Props) {
  const common = {
    width: 16,
    height: 16,
    viewBox: "0 0 24 24",
    fill: "none",
    stroke: "currentColor",
    strokeWidth: 1.75,
    strokeLinecap: "round" as const,
    strokeLinejoin: "round" as const,
    "aria-hidden": true,
  };

  switch (category) {
    case "movies":
      return (
        <svg {...common}>
          <rect x="3" y="5" width="18" height="14" rx="2" />
          <path d="M7 5v14M17 5v14M3 10h4M3 14h4M17 10h4M17 14h4" />
        </svg>
      );
    case "programming":
      return (
        <svg {...common}>
          <path d="m8 8-4 4 4 4M16 8l4 4-4 4M14 6l-4 12" />
        </svg>
      );
    case "shopping":
      return (
        <svg {...common}>
          <path d="M6 7h15l-1.5 9h-12z" />
          <path d="M6 7 5 4H2" />
          <circle cx="9" cy="19" r="1.5" />
          <circle cx="17" cy="19" r="1.5" />
        </svg>
      );
    case "jobs":
      return (
        <svg {...common}>
          <path d="M8 7V5a2 2 0 0 1 2-2h4a2 2 0 0 1 2 2v2" />
          <rect x="3" y="7" width="18" height="13" rx="2" />
          <path d="M3 12h18" />
        </svg>
      );
    case "services":
      return (
        <svg {...common}>
          <path d="M4 13.5 8.5 18a3 3 0 0 0 4.2 0l6.8-6.8a2.4 2.4 0 0 0-3.4-3.4l-5.4 5.4" />
          <path d="M8 12.5 10.5 15M5.5 9.5l3-3 4 4" />
        </svg>
      );
    case "gaming":
      return (
        <svg {...common}>
          <path d="M8 12h4M10 10v4" />
          <path d="M15 11h.01M17 13h.01" />
          <path d="M6 12a8 8 0 0 1 16 0v2a3 3 0 0 1-3 3H9a3 3 0 0 1-3-3z" />
        </svg>
      );
    case "articles":
      return (
        <svg {...common}>
          <path d="M6 4h9l3 3v13H6z" />
          <path d="M15 4v4h4M8 12h8M8 16h8" />
        </svg>
      );
    case "learning":
      return (
        <svg {...common}>
          <path d="M4 7.5 12 4l8 3.5-8 3.5z" />
          <path d="M6 9.5V16l6 3 6-3V9.5" />
        </svg>
      );
    case "music":
      return (
        <svg {...common}>
          <path d="M9 18V5l10-2v13" />
          <circle cx="7" cy="18" r="2.5" />
          <circle cx="17" cy="16" r="2.5" />
        </svg>
      );
    case "news":
      return (
        <svg {...common}>
          <path d="M5 5h14v14H5z" />
          <path d="M8 9h8M8 12h8M8 15h5" />
        </svg>
      );
    case "design":
      return (
        <svg {...common}>
          <circle cx="12" cy="12" r="3" />
          <path d="M12 3v2M12 19v2M3 12h2M19 12h2M5.6 5.6l1.4 1.4M17 17l1.4 1.4M18.4 5.6 17 7M7 17l-1.4 1.4" />
        </svg>
      );
    case "tools":
      return (
        <svg {...common}>
          <path d="M14 4a4 4 0 0 0-5.2 5.2L4 14l6 6 4.8-4.8A4 4 0 0 0 14 4z" />
          <path d="M14 4l6 6" />
        </svg>
      );
    case "entertainment":
      return (
        <svg {...common}>
          <path d="M12 4l1.2 3.6L17 8.8l-3.6 1.2L12 14l-1.2-3.6L7 8.8l3.6-1.2z" />
          <path d="M5 17l.8 2.4L8 20l-2.2.6L5 23l-.8-2.4L2 20l2.2-.6zM19 15l.6 1.8L21 17l-1.8.6L19 19l-.6-1.8L17 17l1.8-.6z" />
        </svg>
      );
    default:
      return (
        <svg {...common}>
          <path d="M4 7a2 2 0 0 1 2-2h4l2 2h6a2 2 0 0 1 2 2v9a2 2 0 0 1-2 2H6a2 2 0 0 1-2-2z" />
        </svg>
      );
  }
}
