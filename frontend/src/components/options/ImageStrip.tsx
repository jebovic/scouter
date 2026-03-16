// frontend/src/components/options/ImageStrip.tsx
import { useState } from "react";
import { useTranslation } from "react-i18next";
import { type OptionImage } from "../../api/images";
import { ImageLightbox } from "./ImageLightbox";
import styles from "./ImageStrip.module.css";

const MAX_VISIBLE = 4;

interface Props {
  images: OptionImage[];
  loading: boolean;
}

export function ImageStrip({ images, loading }: Props) {
  const { t } = useTranslation();
  const [lightboxIndex, setLightboxIndex] = useState<number | null>(null);

  if (loading && images.length === 0) {
    return (
      <div className={styles.skeleton}>
        {[0, 1, 2].map((i) => (
          <div key={i} className={styles.skeletonThumb} />
        ))}
      </div>
    );
  }

  if (images.length === 0) return null;

  const visible = images.slice(0, MAX_VISIBLE);
  const overflow = images.length - MAX_VISIBLE;

  return (
    <>
      <div className={styles.strip} aria-label={t("options.images.gallery")}>
        {visible.map((img, i) => (
          <button
            key={img.id}
            className={styles.thumbBtn}
            onClick={() => setLightboxIndex(i)}
            aria-label={`${t("options.images.view")} ${i + 1}`}
          >
            <img
              src={img.url}
              alt=""
              className={styles.thumb}
              width={80}
              height={60}
            />
          </button>
        ))}
        {overflow > 0 && (
          <button
            className={styles.overflow}
            onClick={() => setLightboxIndex(MAX_VISIBLE)}
          >
            +{overflow}
          </button>
        )}
      </div>
      {lightboxIndex !== null && (
        <ImageLightbox
          images={images}
          initialIndex={lightboxIndex}
          onClose={() => setLightboxIndex(null)}
        />
      )}
    </>
  );
}
