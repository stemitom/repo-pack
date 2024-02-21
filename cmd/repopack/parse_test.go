package repopack

import "testing"

func Test_parseRepoURL(t *testing.T) {
	type args struct {
		urlStr string
	}
	tests := []struct {
		name           string
		args           args
		wantUser       string
		wantRepository string
		wantRef        string
		wantDir        string
		wantErr        bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotUser, gotRepository, gotRef, gotDir, err := parseRepoURL(tt.args.urlStr)
			if (err != nil) != tt.wantErr {
				t.Errorf("parseRepoURL() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if gotUser != tt.wantUser {
				t.Errorf("parseRepoURL() gotUser = %v, want %v", gotUser, tt.wantUser)
			}
			if gotRepository != tt.wantRepository {
				t.Errorf("parseRepoURL() gotRepository = %v, want %v", gotRepository, tt.wantRepository)
			}
			if gotRef != tt.wantRef {
				t.Errorf("parseRepoURL() gotRef = %v, want %v", gotRef, tt.wantRef)
			}
			if gotDir != tt.wantDir {
				t.Errorf("parseRepoURL() gotDir = %v, want %v", gotDir, tt.wantDir)
			}
		})
	}
}
