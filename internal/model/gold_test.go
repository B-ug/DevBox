package model

import "testing"

func TestGoldDefaults(t *testing.T) {
	a := New()
	if a.Gold.Weight != GoldGram || a.Gold.Currency != GoldCurrencyCNY || a.Gold.Window != GoldWindowYear {
		t.Fatalf("%+v", a.Gold)
	}
	tool, ok := Lookup(ToolGold)
	if !ok || tool.Name != "黄金" || tool.Summary != "查看现货金价和历史走势" {
		t.Fatalf("%+v %v", tool, ok)
	}

	a.UseGold("oz", "USD", "5y")
	if a.Gold.Weight != GoldOunce || a.Gold.Currency != "美元" || a.Gold.Window != GoldWindowFiveYear {
		t.Fatalf("%+v", a.Gold)
	}
	if GoldWeightCode(a.Gold.Weight) != "oz" || GoldCurrencyCode(a.Gold.Currency) != "USD" || GoldWindowCode(a.Gold.Window) != "5y" {
		t.Fatalf("codes %+v", a.Gold)
	}

	a.UseGold("lb", "XYZ", "1d")
	if a.Gold.Weight != GoldGram || a.Gold.Currency != GoldCurrencyCNY || a.Gold.Window != GoldWindowYear {
		t.Fatalf("unknown %+v", a.Gold)
	}

	a.UseGold("", "", "")
	if a.Gold.Weight != GoldGram || a.Gold.Currency != GoldCurrencyCNY || a.Gold.Window != GoldWindowYear {
		t.Fatalf("empty %+v", a.Gold)
	}
}
