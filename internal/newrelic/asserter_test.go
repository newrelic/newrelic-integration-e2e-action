package newrelic

import "testing"

func Test_checkBounds(t *testing.T) {
	lowerResult := 5.0
	upperResult := 15.0

	type args struct {
		actualResult        float64
		expectedLowerResult *float64
		expectedUpperResult *float64
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			name:    "when the actual result is below of a dual-bound it should return an error",
			args:    args{actualResult: 2.0, expectedLowerResult: &lowerResult, expectedUpperResult: &upperResult},
			wantErr: true,
		},
		{
			name:    "when the actual result is above of a dual-bound it should return an error",
			args:    args{actualResult: 20.0, expectedLowerResult: &lowerResult, expectedUpperResult: &upperResult},
			wantErr: true,
		},
		{
			name:    "when the actual result is below of an lower bound it should return an error",
			args:    args{actualResult: 1.0, expectedLowerResult: &lowerResult, expectedUpperResult: nil},
			wantErr: true,
		},
		{
			name:    "when the actual result is above of an upper bound it should return an error",
			args:    args{actualResult: 16.0, expectedLowerResult: nil, expectedUpperResult: &upperResult},
			wantErr: true,
		},
		{
			name:    "when the actual result is above of an lower bound it should return no error",
			args:    args{actualResult: 6.0, expectedLowerResult: &lowerResult, expectedUpperResult: nil},
			wantErr: false,
		},
		{
			name:    "when the actual result is equal to a lower bound it should return no error",
			args:    args{actualResult: 5.0, expectedLowerResult: &lowerResult, expectedUpperResult: nil},
			wantErr: false,
		},
		{
			name:    "when the actual result is below of an upper bound it should return no error",
			args:    args{actualResult: 14.0, expectedLowerResult: nil, expectedUpperResult: &upperResult},
			wantErr: false,
		},
		{
			name:    "when the actual result is equal to an upper bound it should return no error",
			args:    args{actualResult: 15.0, expectedLowerResult: nil, expectedUpperResult: &upperResult},
			wantErr: false,
		},
		{
			name:    "when the actual result is between a dual-bound it should return no error",
			args:    args{actualResult: 10.0, expectedLowerResult: &lowerResult, expectedUpperResult: &upperResult},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := checkBounds(tt.args.actualResult, tt.args.expectedLowerResult, tt.args.expectedUpperResult); (err != nil) != tt.wantErr {
				t.Errorf("checkBounds() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
