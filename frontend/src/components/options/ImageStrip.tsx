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
      <div className={styles.strip} role="list" aria-label={t("options.images.gallery")}>
        {visible.map((img, i) => (
          <img
            key={img.id}
            src={img.url}
            alt=""
            className={styles.thumb}
            width={80}
            height={60}
            onClick={() => setLightboxIndex(i)}
          />
        ))}
        {overflow > 0 && (
          <div
            className={styles.overflow}
            onClick={() => setLightboxIndex(MAX_VISIBLE - 1)}
          >
            +{overflow}
          </div>
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
