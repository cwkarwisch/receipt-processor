package main

import "testing"

func TestNamePoints(t *testing.T) {
	t.Run("handles non-alphanumeric characters", func(t *testing.T) {
		got := namePoints("Test!")
		want := 4

		if got != want {
			t.Errorf("expected points of %d but got %d", want, got)
		}
	})

	t.Run("handles letters and digits", func(t *testing.T) {
		got := namePoints("Hi4-.!")
		want := 3

		assertExpectedPoints(t, got, want)
	})
}

func TestRoundDollarPoints(t *testing.T) {
	t.Run("even dollar amount", func(t *testing.T) {
		got := roundDollarPoints("3.00")
		assertExpectedPoints(t, got, 50)
	})

	t.Run("dollars and cents", func(t *testing.T) {
		got := roundDollarPoints("3.21")
		assertExpectedPoints(t, got, 0)
	})
}

func TestMultiplesOfQuartersPoints(t *testing.T) {
	t.Run("not divisible by 0.25", func(t *testing.T) {
		got := multiplesOfQuartersPoints("3.24")
		assertExpectedPoints(t, got, 0)
	})

	t.Run("divisible by 0.25", func(t *testing.T) {
		got := multiplesOfQuartersPoints("3.25")
		assertExpectedPoints(t, got, 25)
	})

	t.Run("divisible by 0.25", func(t *testing.T) {
		got := multiplesOfQuartersPoints("4.00")
		assertExpectedPoints(t, got, 25)
	})

	t.Run("divisible by 0.25", func(t *testing.T) {
		got := multiplesOfQuartersPoints("0.25")
		assertExpectedPoints(t, got, 25)
	})
}

func TestItemPairPoints(t *testing.T) {
	t.Run("single item", func(t *testing.T) {
		got := itemPairPoints(1)
		assertExpectedPoints(t, got, 0)
	})

	t.Run("two items", func(t *testing.T) {
		got := itemPairPoints(2)
		assertExpectedPoints(t, got, 5)
	})

	t.Run("five items", func(t *testing.T) {
		got := itemPairPoints(5)
		assertExpectedPoints(t, got, 10)
	})

	t.Run("seven items", func(t *testing.T) {
		got := itemPairPoints(7)
		assertExpectedPoints(t, got, 15)
	})
}

func TestItemDescriptionPoints(t *testing.T) {
	t.Run("trimmed length is multiple of 3", func(t *testing.T) {
		item := Item{
			ShortDescription: "   Klarbrunn 12-PK 12 FL OZ  ",
			Price:            "12.00",
		}
		got := itemDescriptionPoints(item)
		assertExpectedPoints(t, got, 3)
	})

	t.Run("trimmed length is not a multiple of 3", func(t *testing.T) {
		item := Item{
			ShortDescription: "   larbrunn 12-PK 12 FL OZ  ",
			Price:            "12.00",
		}
		got := itemDescriptionPoints(item)
		assertExpectedPoints(t, got, 0)
	})

	t.Run("trimmed length is a multiple of 3 and total does not need to be rounded after multiplying by 0.2", func(t *testing.T) {
		item := Item{
			ShortDescription: "   Klarbrunn 12-PK 12 FL OZ  ",
			Price:            "10.00",
		}
		got := itemDescriptionPoints(item)
		assertExpectedPoints(t, got, 2)
	})
}

func TestItemPoints(t *testing.T) {
	// using same items as above, checking that the points are summed up correctly for the collection
	item1 := Item{
		ShortDescription: "   Klarbrunn 12-PK 12 FL OZ  ",
		Price:            "12.00",
	}
	item2 := Item{
		ShortDescription: "   larbrunn 12-PK 12 FL OZ  ",
		Price:            "12.00",
	}
	item3 := Item{
		ShortDescription: "   Klarbrunn 12-PK 12 FL OZ  ",
		Price:            "10.00",
	}
	items := []Item{item1, item2, item3}
	got := itemPoints(items)
	assertExpectedPoints(t, got, 5)
}

func TestPurchaseDatePoints(t *testing.T) {
	t.Run("date is even", func(t *testing.T) {
		dateString := "2022-01-28"
		got := purchaseDatePoints(dateString)
		assertExpectedPoints(t, got, 0)
	})

	t.Run("date is odd", func(t *testing.T) {
		dateString := "2022-01-31"
		got := purchaseDatePoints(dateString)
		assertExpectedPoints(t, got, 6)
	})
}

func TestPurchaseTimePoints(t *testing.T) {
	t.Run("purchased at 2:00pm", func(t *testing.T) {
		time := "14:00"
		got := purchaseTimePoints(time)
		assertExpectedPoints(t, got, 0)
	})

	t.Run("purchased at 4:00pm", func(t *testing.T) {
		time := "16:00"
		got := purchaseTimePoints(time)
		assertExpectedPoints(t, got, 0)
	})

	t.Run("purchased at 2:01pm", func(t *testing.T) {
		time := "14:01"
		got := purchaseTimePoints(time)
		assertExpectedPoints(t, got, 10)
	})

	t.Run("purchased at 3:59pm", func(t *testing.T) {
		time := "15:59"
		got := purchaseTimePoints(time)
		assertExpectedPoints(t, got, 10)
	})
}

func assertExpectedPoints(t testing.TB, got, want int) {
	t.Helper()
	if got != want {
		t.Errorf("expected points of %d but got %d", want, got)
	}
}
