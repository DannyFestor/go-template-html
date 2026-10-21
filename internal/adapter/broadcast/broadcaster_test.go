package broadcast_test

import (
	"context"
	"errors"
	"fmt"
	"io"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/DannyFestor/go-template-html/internal/adapter/broadcast"
	"github.com/DannyFestor/go-template-html/internal/domain"
)

var (
	errRender  = errors.New("render failed")
	errPublish = errors.New("valkey is down")
)

type published struct {
	channel string
	payload string
}

type fakePublisher struct {
	published []published
	err       error
}

func (p *fakePublisher) Publish(_ context.Context, channel string, payload []byte) error {
	if p.err != nil {
		return p.err
	}

	p.published = append(p.published, published{channel: channel, payload: string(payload)})

	return nil
}

type spyReporter struct {
	reported []error
}

func (r *spyReporter) Report(_ context.Context, err error) {
	r.reported = append(r.reported, err)
}

type textFragment struct {
	text string
	err  error
}

func (f textFragment) Render(_ context.Context, w io.Writer) error {
	if f.err != nil {
		return f.err
	}

	_, err := io.WriteString(w, "<p>"+f.text+"</p>")
	if err != nil {
		return fmt.Errorf("write the fragment: %w", err)
	}

	return nil
}

func composeAnnouncement(announcement domain.Announcement) broadcast.Message {
	return broadcast.Message{Channel: "announcements", Fragment: textFragment{text: announcement.Text, err: nil}}
}

func TestABroadcastIsRenderedAndPublishedOnItsChannel(t *testing.T) {
	t.Parallel()

	publisher := &fakePublisher{published: nil, err: nil}
	reporter := &spyReporter{reported: nil}

	broadcast.NewBroadcaster(composeAnnouncement, publisher, reporter).
		Broadcast(t.Context(), domain.Announcement{Text: "Hello"})

	assert.Equal(t, []published{{channel: "announcements", payload: "<p>Hello</p>"}}, publisher.published)
	assert.Empty(t, reporter.reported)
}

func TestAFragmentThatFailsToRenderIsReportedAndNotPublished(t *testing.T) {
	t.Parallel()

	publisher := &fakePublisher{published: nil, err: nil}
	reporter := &spyReporter{reported: nil}
	failing := func(domain.Announcement) broadcast.Message {
		return broadcast.Message{Channel: "announcements", Fragment: textFragment{text: "", err: errRender}}
	}

	broadcast.NewBroadcaster(failing, publisher, reporter).Broadcast(t.Context(), domain.Announcement{Text: "Hello"})

	assert.Empty(t, publisher.published)
	assert.Len(t, reporter.reported, 1)
	assert.ErrorIs(t, reporter.reported[0], errRender)
}

func TestAFailedPublishIsReportedAndSwallowed(t *testing.T) {
	t.Parallel()

	reporter := &spyReporter{reported: nil}

	broadcast.NewBroadcaster(composeAnnouncement, &fakePublisher{published: nil, err: errPublish}, reporter).
		Broadcast(t.Context(), domain.Announcement{Text: "Hello"})

	assert.Len(t, reporter.reported, 1)
	assert.ErrorIs(t, reporter.reported[0], errPublish)
}
