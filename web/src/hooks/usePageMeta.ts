import { useEffect } from "react";
import { OG_IMAGE, SITE_NAME, SITE_URL } from "../seo/site";

type PageMeta = {
  title: string;
  description: string;
  path?: string;
  noindex?: boolean;
};

function upsertMeta(
  selector: string,
  create: () => HTMLMetaElement | HTMLLinkElement,
  apply: (element: HTMLMetaElement | HTMLLinkElement) => void,
) {
  let element = document.head.querySelector(selector) as
    | HTMLMetaElement
    | HTMLLinkElement
    | null;

  if (!element) {
    element = create();
    document.head.appendChild(element);
  }

  apply(element);
}

export function usePageMeta({ title, description, path = "/", noindex = false }: PageMeta) {
  useEffect(() => {
    const canonical = `${SITE_URL}${path}`;

    document.title = title;

    upsertMeta(
      'meta[name="description"]',
      () => {
        const meta = document.createElement("meta");
        meta.setAttribute("name", "description");
        return meta;
      },
      (meta) => meta.setAttribute("content", description),
    );

    upsertMeta(
      'meta[name="robots"]',
      () => {
        const meta = document.createElement("meta");
        meta.setAttribute("name", "robots");
        return meta;
      },
      (meta) =>
        meta.setAttribute("content", noindex ? "noindex, nofollow" : "index, follow"),
    );

    upsertMeta(
      'link[rel="canonical"]',
      () => {
        const link = document.createElement("link");
        link.setAttribute("rel", "canonical");
        return link;
      },
      (link) => link.setAttribute("href", canonical),
    );

    const ogTags: Array<[string, string]> = [
      ["og:type", "website"],
      ["og:site_name", SITE_NAME],
      ["og:title", title],
      ["og:description", description],
      ["og:url", canonical],
      ["og:image", OG_IMAGE],
      ["og:locale", "ru_RU"],
    ];

    for (const [property, content] of ogTags) {
      upsertMeta(
        `meta[property="${property}"]`,
        () => {
          const meta = document.createElement("meta");
          meta.setAttribute("property", property);
          return meta;
        },
        (meta) => meta.setAttribute("content", content),
      );
    }

    const twitterTags: Array<[string, string]> = [
      ["twitter:card", "summary"],
      ["twitter:title", title],
      ["twitter:description", description],
      ["twitter:image", OG_IMAGE],
    ];

    for (const [name, content] of twitterTags) {
      upsertMeta(
        `meta[name="${name}"]`,
        () => {
          const meta = document.createElement("meta");
          meta.setAttribute("name", name);
          return meta;
        },
        (meta) => meta.setAttribute("content", content),
      );
    }
  }, [title, description, path, noindex]);
}
