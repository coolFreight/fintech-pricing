package pricing

import (
	"testing"
)

var _ = func() bool {
	testing.Init()
	return true
}()

func TestPricingSimulator1(t *testing.T) {
	//	RandomizeStart("FAKEPACA")
	//	var wg sync.WaitGroup
	//	wg.Add(1)
	//	go func() {
	//		defer wg.Done()
	//		time.Sleep(100 * time.Second)
	//	}()
	//	wg.Wait()
}

func TestPricingSimulator(t *testing.T) {
	//var logger = slog.New(slog.NewJSONHandler(os.Stdout, nil))
	//url, _, err := RandomizeStart("FAKEPACA")
	//os.Setenv(internal.APCA_MARKET_PRICING_STREAM, url)
	//os.Setenv(internal.APCA_BASE_URL, url)
	//
	//if err != nil {
	//	t.Errorf("Error creating priceSimulator: %v", err)
	//}
	//
	//conn, err := internal.Connect()
	//if err != nil {
	//	slog.Error("Error connecting to pricing", slog.Any("error", err))
	//	return
	//}
	//pc := NewPricingClient(conn, logger)
	//
	//ch, _ := pc.Start([]string{"FAKEPACA"})
	////var wg sync.WaitGroup
	////wg.Add(1)
	////go func() {
	////defer wg.Done()
	//for prices := range ch {
	//	for _, price := range prices {
	//		t.Logf("Received price %v", price)
	//	}
	//	//return
	//}
	////}()
	////wg.Wait()
	//t.Log("All done with tests")
}
