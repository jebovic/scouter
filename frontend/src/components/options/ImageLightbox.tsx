import { useState, useEffect } from "react";
import { useTranslation } from "react-i18next";
import { type OptionImage } from "../../api/images";
import styles from "./ImageLightbox.module.css";

interface Props {
  images: OptionImage[];
  initialIndex: number;
  onClose: () => void;
}

export function ImageLightbox({ images, initialIndex, onClose }: Props) {
  const { t } = useTranslation();
  const [index, setIndex] = useState(initialIndex);

  useEffect(() => {
    const onKey = (e: KeyboardEvent) => {
      if (e.key === "Escape") onClose();
      if (e.key === "ArrowLeft") setIndex((i) => Math.max(0, i - 1));
      if (e.key === "ArrowRight") setIndex((i) => Math.min(images.length - 1, i + 1));
    };
    window.addEventListener("keydown", onKey);
    return () => window.removeEventListener("keydown", onKey);
  }, [images.length, onClose]);

  const img = images[index];

  return (
    <div className={styles.overlay} onClick={onClose} role="dialog" aria-modal="true">
      <button
        className={styles.close}
        onClick={onClose}
        aria-label={t("options.images.lightboxClose")}
      >
        ×
      </button>
      <img
        src={img.url}
        alt=""
        className={styles.image}
        onClick={(e) => e.stopPropagation()}
      />
      {images.length > 1 && (
        <>
          <button
            className={`${styles.nav} ${styles.navPrev}`}
            onClick={(e) => { e.stopPropagation(); setIndex((i) => Math.max(0, i - 1)); }}
            disabled={index === 0}
            aria-label={t("options.images.prevImage")}
          >
            ‹
          </button>
          <button
            className={`${styles.nav} ${styles.navNext}`}
            onClick={(e) => { e.stopPropagation(); setIndex((i) => Math.min(images.length - 1, i + 1)); }}
            disabled={index === images.length - 1}
            aria-label={t("options.images.nextImage")}
          >
            ›
          </button>
        </>
      )}
    </div>
  );
}
