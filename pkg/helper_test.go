package pkg

import "testing"

func TestDomain_Check(t *testing.T) {
	tests := []struct {
		name string
		d    *Domain
		want bool
	}{
		{
			name: "www.shoppersstop.com",
			d: &Domain{
				Name: "www.shoppersstop.com",
				Target: []string{
					"k.sni.global.fastly.net",
					"152.199.5.41",
					"www-shoppersstop-com-1edf.fast.getn7.io.",
				},
			},
			want: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.d.Check(); got != tt.want {
				t.Errorf("Domain.Check() = %v, want %v", got, tt.want)
			}
		})
	}
}
