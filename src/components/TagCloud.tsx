import { useEffect, useState } from "react";
import { Link, useSearchParams } from "react-router-dom";
import { fetchTags, type TagItem } from "@/data/videos";

export function TagCloud() {
  const [params] = useSearchParams();
  const activeTag = params.get("tag");
  const [tags, setTags] = useState<TagItem[]>([]);

  useEffect(() => {
    let active = true;
    fetchTags().then((list) => {
      if (active) setTags(list);
    });
    return () => {
      active = false;
    };
  }, []);

  if (tags.length === 0) return null;

  // 将标签分为奇偶两行，使其横向自由流式排布，不发生强制的列对齐
  const row1 = tags.filter((_, idx) => idx % 2 === 0);
  const row2 = tags.filter((_, idx) => idx % 2 !== 0);

  const renderTag = (tag: TagItem) => (
    <Link
      key={tag.id}
      to={`/list?tag=${encodeURIComponent(tag.label)}`}
      className={`tag-chip ${activeTag === tag.label ? "is-active" : ""}`}
      title={
        typeof tag.count === "number" ? `${tag.count} 个视频` : undefined
      }
    >
      {tag.label}
      {typeof tag.count === "number" && tag.count > 0 && (
        <span style={{ marginLeft: 4, opacity: 0.7 }}>({tag.count})</span>
      )}
    </Link>
  );

  return (
    <div className="tag-cloud-container" aria-label="热门分类">
      <div className="tag-cloud__grid">
        <div className="tag-cloud__row">
          {row1.map(renderTag)}
        </div>
        <div className="tag-cloud__row">
          {row2.map(renderTag)}
        </div>
      </div>
    </div>
  );
}
