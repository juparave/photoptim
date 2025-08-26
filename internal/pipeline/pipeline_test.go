package pipeline

import (
	"context"
	"image"
	"image/color"
	"image/jpeg"
	"testing"
	"time"

	"photoptim/internal/optimizer"
	"photoptim/internal/remotefs"
)

func genJPEG() []byte {
	img := image.NewRGBA(image.Rect(0, 0, 8, 8))
	for y := 0; y < 8; y++ {
		for x := 0; x < 8; x++ {
			img.Set(x, y, color.RGBA{uint8(x * 32), uint8(y * 32), 0, 255})
		}
	}
	var buf []byte
	w := &sliceWriter{&buf}
	_ = jpeg.Encode(w, img, &jpeg.Options{Quality: 80})
	return buf
}

type sliceWriter struct{ b *[]byte }

func (w *sliceWriter) Write(p []byte) (int, error) { *w.b = append(*w.b, p...); return len(p), nil }

func TestOrchestratorRun(t *testing.T) {
	fs := remotefs.NewMockFS("/")
	img1 := genJPEG()
	img2 := genJPEG()
	fs.PutTestFile("/a.jpg", img1)
	fs.PutTestFile("/b.jpg", img2)
	opt := optimizer.New()
	orch := Orchestrator{FS: fs, Opt: opt, Concurrency: 2, JPEGQuality: 75}
	tasks := []FileTask{{Entry: remotefs.RemoteEntry{Path: "/a.jpg", Name: "a.jpg", Size: int64(len(img1))}}, {Entry: remotefs.RemoteEntry{Path: "/b.jpg", Name: "b.jpg", Size: int64(len(img2))}}}
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	prog, errs := orch.Run(ctx, tasks)
	var dl, optc, up int
	for ev := range prog {
		if ev.Err != nil {
			t.Fatalf("unexpected error: %v", ev.Err)
		}
		switch ev.Phase {
		case PhaseDownload:
			dl++
		case PhaseOptimize:
			optc++
		case PhaseUpload:
			up++
		}
	}
	select {
	case e := <-errs:
		if e != nil {
			t.Fatalf("errs: %v", e)
		}
	default:
	}
	if dl != 2 || optc != 2 || up != 2 {
		t.Fatalf("phase counts mismatch dl=%d opt=%d up=%d", dl, optc, up)
	}
}
