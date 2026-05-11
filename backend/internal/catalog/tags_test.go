package catalog

import (
	"context"
	"testing"
	"time"
)

func TestCreateTagAndClassifyAddsTagToMatchingExistingVideos(t *testing.T) {
	ctx := context.Background()
	cat, err := Open(t.TempDir() + "/catalog.db")
	if err != nil {
		t.Fatalf("open catalog: %v", err)
	}
	t.Cleanup(func() {
		if err := cat.Close(); err != nil {
			t.Fatalf("close catalog: %v", err)
		}
	})

	now := time.Now()
	if err := cat.UpsertVideo(ctx, &Video{
		ID:          "video-1",
		DriveID:     "drive",
		FileID:      "file-1",
		Title:       "清纯短发合集",
		Category:    "普通目录",
		PublishedAt: now,
		CreatedAt:   now,
		UpdatedAt:   now,
	}); err != nil {
		t.Fatalf("seed matching video: %v", err)
	}
	if err := cat.UpsertVideo(ctx, &Video{
		ID:          "video-2",
		DriveID:     "drive",
		FileID:      "file-2",
		Title:       "普通标题",
		Category:    "普通目录",
		PublishedAt: now,
		CreatedAt:   now,
		UpdatedAt:   now,
	}); err != nil {
		t.Fatalf("seed non-matching video: %v", err)
	}

	classified, err := cat.CreateTagAndClassify(ctx, "清纯", nil, "user")
	if err != nil {
		t.Fatalf("create tag: %v", err)
	}
	if classified != 1 {
		t.Fatalf("classified = %d, want 1", classified)
	}

	got, err := cat.GetVideo(ctx, "video-1")
	if err != nil {
		t.Fatalf("get matching video: %v", err)
	}
	if !sameStrings(got.Tags, []string{"清纯"}) {
		t.Fatalf("matching tags = %#v, want 清纯", got.Tags)
	}

	other, err := cat.GetVideo(ctx, "video-2")
	if err != nil {
		t.Fatalf("get non-matching video: %v", err)
	}
	if len(other.Tags) != 0 {
		t.Fatalf("non-matching tags = %#v, want none", other.Tags)
	}
}

func TestSetManualVideoTagsRejectsUnknownLabels(t *testing.T) {
	ctx := context.Background()
	cat, err := Open(t.TempDir() + "/catalog.db")
	if err != nil {
		t.Fatalf("open catalog: %v", err)
	}
	t.Cleanup(func() {
		if err := cat.Close(); err != nil {
			t.Fatalf("close catalog: %v", err)
		}
	})

	now := time.Now()
	if err := cat.UpsertVideo(ctx, &Video{
		ID:          "video-1",
		DriveID:     "drive",
		FileID:      "file-1",
		Title:       "普通标题",
		PublishedAt: now,
		CreatedAt:   now,
		UpdatedAt:   now,
	}); err != nil {
		t.Fatalf("seed video: %v", err)
	}

	if err := cat.SetManualVideoTags(ctx, "video-1", []string{"不存在"}); err == nil {
		t.Fatal("SetManualVideoTags accepted an unknown tag label")
	}
}

func TestSetAutoVideoTagsDoesNotOverwriteManualVideoTags(t *testing.T) {
	ctx := context.Background()
	cat, err := Open(t.TempDir() + "/catalog.db")
	if err != nil {
		t.Fatalf("open catalog: %v", err)
	}
	t.Cleanup(func() {
		if err := cat.Close(); err != nil {
			t.Fatalf("close catalog: %v", err)
		}
	})

	now := time.Now()
	if err := cat.UpsertVideo(ctx, &Video{
		ID:          "video-1",
		DriveID:     "drive",
		FileID:      "file-1",
		Title:       "清纯后入",
		PublishedAt: now,
		CreatedAt:   now,
		UpdatedAt:   now,
	}); err != nil {
		t.Fatalf("seed video: %v", err)
	}
	if _, err := cat.CreateTagAndClassify(ctx, "清纯", nil, "user"); err != nil {
		t.Fatalf("create user tag: %v", err)
	}
	if err := cat.SetManualVideoTags(ctx, "video-1", []string{"清纯"}); err != nil {
		t.Fatalf("set manual tags: %v", err)
	}

	if err := cat.SetAutoVideoTags(ctx, "video-1", []string{"后入"}); err != nil {
		t.Fatalf("set auto tags: %v", err)
	}

	got, err := cat.GetVideo(ctx, "video-1")
	if err != nil {
		t.Fatalf("get video: %v", err)
	}
	if !sameStrings(got.Tags, []string{"清纯"}) {
		t.Fatalf("tags = %#v, want manual 清纯 only", got.Tags)
	}
}

func TestCreateTagAndClassifyMapsAVCodeLabelToAV(t *testing.T) {
	ctx := context.Background()
	cat, err := Open(t.TempDir() + "/catalog.db")
	if err != nil {
		t.Fatalf("open catalog: %v", err)
	}
	t.Cleanup(func() {
		if err := cat.Close(); err != nil {
			t.Fatalf("close catalog: %v", err)
		}
	})

	if _, err := cat.CreateTagAndClassify(ctx, "cc-1750027", nil, "user"); err != nil {
		t.Fatalf("create code tag: %v", err)
	}

	tags, err := cat.ListTags(ctx)
	if err != nil {
		t.Fatalf("list tags: %v", err)
	}
	for _, tag := range tags {
		if tag.Label == "cc-1750027" {
			t.Fatal("created standalone AV code tag cc-1750027")
		}
	}
}

func TestLooksLikeCollectionTagRejectsAVCodes(t *testing.T) {
	cases := []string{
		"DASS-499-C",
		"dass-499-c",
		"ADN-778",
		"SONE-247-C",
		"JUQ-502-UC",
		"ABF-032",
		"SSIS-233",
		"MIDA-607",
		"cc-1750027",
		"FC2-PPV-74663555",
		"ADN-778-FHD(1)",
		"ADN-778-中文字幕",
		"[44x.me]idbd-786",
		"NTRH-018_FHD_CH",
		"390JAC-233",
	}
	for _, label := range cases {
		if LooksLikeCollectionTag(label) {
			t.Fatalf("LooksLikeCollectionTag(%q) = true, want false", label)
		}
	}
}

func TestMigrateCollapsesAVCodeTagsIntoAV(t *testing.T) {
	ctx := context.Background()
	cat, err := Open(t.TempDir() + "/catalog.db")
	if err != nil {
		t.Fatalf("open catalog: %v", err)
	}
	t.Cleanup(func() {
		if err := cat.Close(); err != nil {
			t.Fatalf("close catalog: %v", err)
		}
	})

	now := time.Now()
	for _, seed := range []struct {
		id    string
		label string
	}{
		{id: "video-1", label: "cc-1750027"},
		{id: "video-2", label: "ADN-778-FHD(1)"},
		{id: "video-3", label: "[44x.me]idbd-786"},
		{id: "video-4", label: "390JAC-233"},
	} {
		if err := cat.UpsertVideo(ctx, &Video{
			ID:          seed.id,
			DriveID:     "drive",
			FileID:      seed.id,
			Title:       seed.label + " sample",
			Tags:        []string{seed.label},
			Category:    seed.label,
			PublishedAt: now,
			CreatedAt:   now,
			UpdatedAt:   now,
		}); err != nil {
			t.Fatalf("seed polluted video %s: %v", seed.label, err)
		}
	}

	if err := cat.migrate(ctx); err != nil {
		t.Fatalf("migrate: %v", err)
	}

	tags, err := cat.ListTags(ctx)
	if err != nil {
		t.Fatalf("list tags: %v", err)
	}
	var sawAV bool
	polluted := map[string]bool{}
	for _, tag := range tags {
		if tag.Label == "AV" {
			sawAV = true
		}
		if tag.Label != "AV" && isAVCodePollutedLabel(tag.Label) {
			polluted[tag.Label] = true
		}
	}
	if !sawAV {
		t.Fatal("AV tag was not seeded")
	}
	if len(polluted) > 0 {
		t.Fatalf("AV code tags were not removed: %#v", polluted)
	}

	for _, id := range []string{"video-1", "video-2", "video-3", "video-4"} {
		got, err := cat.GetVideo(ctx, id)
		if err != nil {
			t.Fatalf("get video %s: %v", id, err)
		}
		if !sameStrings(got.Tags, []string{"AV"}) {
			t.Fatalf("%s tags = %#v, want AV", id, got.Tags)
		}
	}
}

func TestListVideosHidesDuplicateContentHashes(t *testing.T) {
	ctx := context.Background()
	cat, err := Open(t.TempDir() + "/catalog.db")
	if err != nil {
		t.Fatalf("open catalog: %v", err)
	}
	t.Cleanup(func() {
		if err := cat.Close(); err != nil {
			t.Fatalf("close catalog: %v", err)
		}
	})

	now := time.Now()
	for _, v := range []*Video{
		{
			ID:          "video-1",
			DriveID:     "drive",
			FileID:      "file-1",
			Title:       "First",
			ContentHash: "hash-same",
			PublishedAt: now,
			CreatedAt:   now,
			UpdatedAt:   now,
		},
		{
			ID:          "video-2",
			DriveID:     "drive",
			FileID:      "file-2",
			Title:       "Second",
			ContentHash: "hash-same",
			PublishedAt: now.Add(time.Second),
			CreatedAt:   now.Add(time.Second),
			UpdatedAt:   now.Add(time.Second),
		},
	} {
		if err := cat.UpsertVideo(ctx, v); err != nil {
			t.Fatalf("seed video %s: %v", v.ID, err)
		}
	}

	items, total, err := cat.ListVideos(ctx, ListParams{Page: 1, PageSize: 10})
	if err != nil {
		t.Fatalf("list videos: %v", err)
	}
	if total != 1 || len(items) != 1 {
		t.Fatalf("visible videos total=%d len=%d, want 1", total, len(items))
	}
	if items[0].ID != "video-1" {
		t.Fatalf("visible id = %q, want video-1", items[0].ID)
	}
}

func sameStrings(a, b []string) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}
