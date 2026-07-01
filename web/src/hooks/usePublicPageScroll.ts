import { useEffect } from "react";

export function usePublicPageScroll() {
  useEffect(() => {
    document.documentElement.classList.add("is-public-page");
    document.body.classList.add("is-public-page");

    return () => {
      document.documentElement.classList.remove("is-public-page");
      document.body.classList.remove("is-public-page");
    };
  }, []);
}
