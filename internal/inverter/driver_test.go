package inverter

import (
	"testing"

	"go.uber.org/zap"
)

type fakeDriver struct{ id string }

func (f *fakeDriver) Model() string { return "fake" }
func (f *fakeDriver) DeviceID() string { return f.id }
func (f *fakeDriver) Scan(ctx context.Context) (Snapshot, error) {
	return Snapshot{Model: "fake", DeviceID: f.id, Status: "ok"}, nil
}
func (f *fakeDriver) Close() error { return nil }

func TestSnapshotShape(t *testing.T) {
	d := &fakeDriver{id: "unit-1"}
	s, err := d.Scan(nil)
	if err != nil {
		t.Fatal(err)
	}
	if s.Model != "fake" || s.DeviceID != "unit-1" || s.Status != "ok" {
		t.Fatalf("unexpected snapshot: %+v", s)
	}
}

var _ Driver = (*fakeDriver)(nil)
var _ = zap.NewNop
