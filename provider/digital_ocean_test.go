/*
Copyright 2017 The Kubernetes Authors.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package provider

import (
	"fmt"
	"os"
	"testing"

	"github.com/digitalocean/godo"
	"github.com/digitalocean/godo/context"

	"github.com/kubernetes-incubator/external-dns/endpoint"
	"github.com/kubernetes-incubator/external-dns/plan"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type mockDigitalOceanInterface interface {
	List(context.Context, *godo.ListOptions) ([]godo.Domain, *godo.Response, error)
	Create(context.Context, *godo.DomainCreateRequest) (*godo.Domain, *godo.Response, error)
	Delete(context.Context, string) (*godo.Response, error)
}

type mockDigitalOceanClient struct{}

func (m *mockDigitalOceanClient) List(ctx context.Context, opt *godo.ListOptions) ([]godo.Domain, *godo.Response, error) {
	if opt == nil || opt.Page == 0 {
		return []godo.Domain{{Name: "foo.com"}}, &godo.Response{
			Links: &godo.Links{
				Pages: &godo.Pages{
					Next: "http://example.com/v2/domains/?page=2",
					Last: "1234",
				},
			},
		}, nil
	}
	return []godo.Domain{{Name: "bar.com"}, {Name: "bar.de"}}, nil, nil
}

func (m *mockDigitalOceanClient) Create(context.Context, *godo.DomainCreateRequest) (*godo.Domain, *godo.Response, error) {
	return &godo.Domain{Name: "example.com"}, nil, nil
}

func (m *mockDigitalOceanClient) CreateRecord(context.Context, string, *godo.DomainRecordEditRequest) (*godo.DomainRecord, *godo.Response, error) {
	return &godo.DomainRecord{ID: 1}, nil, nil
}

func (m *mockDigitalOceanClient) Delete(context.Context, string) (*godo.Response, error) {
	return nil, nil
}
func (m *mockDigitalOceanClient) DeleteRecord(ctx context.Context, domain string, id int) (*godo.Response, error) {
	return nil, nil
}
func (m *mockDigitalOceanClient) EditRecord(ctx context.Context, domain string, id int, editRequest *godo.DomainRecordEditRequest) (*godo.DomainRecord, *godo.Response, error) {
	return &godo.DomainRecord{ID: 1}, nil, nil
}

func (m *mockDigitalOceanClient) Get(ctx context.Context, name string) (*godo.Domain, *godo.Response, error) {
	return &godo.Domain{Name: "example.com"}, nil, nil
}

func (m *mockDigitalOceanClient) Record(ctx context.Context, domain string, id int) (*godo.DomainRecord, *godo.Response, error) {
	return &godo.DomainRecord{ID: 1}, nil, nil
}

func (m *mockDigitalOceanClient) Records(ctx context.Context, domain string, opt *godo.ListOptions) ([]godo.DomainRecord, *godo.Response, error) {
	if domain == "foo.com" {
		if opt == nil || opt.Page == 0 {
			return []godo.DomainRecord{{ID: 1, Name: "foobar.ext-dns-test.foo.com."}, {ID: 2}}, &godo.Response{
				Links: &godo.Links{
					Pages: &godo.Pages{
						Next: "http://example.com/v2/domains/?page=2",
						Last: "1234",
					},
				},
			}, nil
		}
		return []godo.DomainRecord{{ID: 3, Name: "baz.ext-dns-test.foo.com."}}, nil, nil
	}
	return nil, nil, nil
}

type mockDigitalOceanListFail struct{}

func (m *mockDigitalOceanListFail) List(context.Context, *godo.ListOptions) ([]godo.Domain, *godo.Response, error) {
	return []godo.Domain{}, nil, fmt.Errorf("Fail to get domains")
}

func (m *mockDigitalOceanListFail) Create(context.Context, *godo.DomainCreateRequest) (*godo.Domain, *godo.Response, error) {
	return &godo.Domain{Name: "example.com"}, nil, nil
}

func (m *mockDigitalOceanListFail) CreateRecord(context.Context, string, *godo.DomainRecordEditRequest) (*godo.DomainRecord, *godo.Response, error) {
	return &godo.DomainRecord{ID: 1}, nil, nil
}

func (m *mockDigitalOceanListFail) Delete(context.Context, string) (*godo.Response, error) {
	return nil, nil
}
func (m *mockDigitalOceanListFail) DeleteRecord(ctx context.Context, domain string, id int) (*godo.Response, error) {
	return nil, nil
}
func (m *mockDigitalOceanListFail) EditRecord(ctx context.Context, domain string, id int, editRequest *godo.DomainRecordEditRequest) (*godo.DomainRecord, *godo.Response, error) {
	return &godo.DomainRecord{ID: 1}, nil, nil
}

func (m *mockDigitalOceanListFail) Get(ctx context.Context, name string) (*godo.Domain, *godo.Response, error) {
	return &godo.Domain{Name: "example.com"}, nil, nil
}

func (m *mockDigitalOceanListFail) Record(ctx context.Context, domain string, id int) (*godo.DomainRecord, *godo.Response, error) {
	return &godo.DomainRecord{ID: 1}, nil, nil
}

func (m *mockDigitalOceanListFail) Records(ctx context.Context, domain string, opt *godo.ListOptions) ([]godo.DomainRecord, *godo.Response, error) {
	return []godo.DomainRecord{{ID: 1}, {ID: 2}}, nil, nil
}

type mockDigitalOceanGetFail struct{}

func (m *mockDigitalOceanGetFail) List(context.Context, *godo.ListOptions) ([]godo.Domain, *godo.Response, error) {
	return []godo.Domain{{Name: "foo.com"}, {Name: "bar.com"}}, nil, nil
}

func (m *mockDigitalOceanGetFail) Create(context.Context, *godo.DomainCreateRequest) (*godo.Domain, *godo.Response, error) {
	return &godo.Domain{Name: "example.com"}, nil, nil
}

func (m *mockDigitalOceanGetFail) CreateRecord(context.Context, string, *godo.DomainRecordEditRequest) (*godo.DomainRecord, *godo.Response, error) {
	return &godo.DomainRecord{ID: 1}, nil, nil
}

func (m *mockDigitalOceanGetFail) Delete(context.Context, string) (*godo.Response, error) {
	return nil, nil
}
func (m *mockDigitalOceanGetFail) DeleteRecord(ctx context.Context, domain string, id int) (*godo.Response, error) {
	return nil, nil
}
func (m *mockDigitalOceanGetFail) EditRecord(ctx context.Context, domain string, id int, editRequest *godo.DomainRecordEditRequest) (*godo.DomainRecord, *godo.Response, error) {
	return &godo.DomainRecord{ID: 1}, nil, nil
}

func (m *mockDigitalOceanGetFail) Get(ctx context.Context, name string) (*godo.Domain, *godo.Response, error) {
	return nil, nil, fmt.Errorf("Failed to get domain")
}

func (m *mockDigitalOceanGetFail) Record(ctx context.Context, domain string, id int) (*godo.DomainRecord, *godo.Response, error) {
	return &godo.DomainRecord{ID: 1}, nil, nil
}

func (m *mockDigitalOceanGetFail) Records(ctx context.Context, domain string, opt *godo.ListOptions) ([]godo.DomainRecord, *godo.Response, error) {
	return []godo.DomainRecord{{ID: 1}, {ID: 2}}, nil, nil
}

type mockDigitalOceanCreateFail struct{}

func (m *mockDigitalOceanCreateFail) List(context.Context, *godo.ListOptions) ([]godo.Domain, *godo.Response, error) {
	return []godo.Domain{{Name: "foo.com"}, {Name: "bar.com"}}, nil, nil
}

func (m *mockDigitalOceanCreateFail) Create(context.Context, *godo.DomainCreateRequest) (*godo.Domain, *godo.Response, error) {
	return nil, nil, fmt.Errorf("Failed to create domain")
}

func (m *mockDigitalOceanCreateFail) CreateRecord(context.Context, string, *godo.DomainRecordEditRequest) (*godo.DomainRecord, *godo.Response, error) {
	return &godo.DomainRecord{ID: 1}, nil, nil
}

func (m *mockDigitalOceanCreateFail) Delete(context.Context, string) (*godo.Response, error) {
	return nil, nil
}
func (m *mockDigitalOceanCreateFail) DeleteRecord(ctx context.Context, domain string, id int) (*godo.Response, error) {
	return nil, nil
}
func (m *mockDigitalOceanCreateFail) EditRecord(ctx context.Context, domain string, id int, editRequest *godo.DomainRecordEditRequest) (*godo.DomainRecord, *godo.Response, error) {
	return &godo.DomainRecord{ID: 1}, nil, nil
}

func (m *mockDigitalOceanCreateFail) Get(ctx context.Context, name string) (*godo.Domain, *godo.Response, error) {
	return &godo.Domain{Name: "example.com"}, nil, nil
}

func (m *mockDigitalOceanCreateFail) Record(ctx context.Context, domain string, id int) (*godo.DomainRecord, *godo.Response, error) {
	return &godo.DomainRecord{ID: 1}, nil, nil
}

func (m *mockDigitalOceanCreateFail) Records(ctx context.Context, domain string, opt *godo.ListOptions) ([]godo.DomainRecord, *godo.Response, error) {
	return []godo.DomainRecord{{ID: 1}, {ID: 2}}, nil, nil
}

type mockDigitalOceanDeleteFail struct{}

func (m *mockDigitalOceanDeleteFail) List(context.Context, *godo.ListOptions) ([]godo.Domain, *godo.Response, error) {
	return []godo.Domain{{Name: "foo.com"}, {Name: "bar.com"}}, nil, nil
}

func (m *mockDigitalOceanDeleteFail) Create(context.Context, *godo.DomainCreateRequest) (*godo.Domain, *godo.Response, error) {
	return &godo.Domain{Name: "example.com"}, nil, nil
}

func (m *mockDigitalOceanDeleteFail) CreateRecord(context.Context, string, *godo.DomainRecordEditRequest) (*godo.DomainRecord, *godo.Response, error) {
	return &godo.DomainRecord{ID: 1}, nil, nil
}

func (m *mockDigitalOceanDeleteFail) Delete(context.Context, string) (*godo.Response, error) {
	return nil, fmt.Errorf("Failed to delete domain")
}
func (m *mockDigitalOceanDeleteFail) DeleteRecord(ctx context.Context, domain string, id int) (*godo.Response, error) {
	return nil, nil
}
func (m *mockDigitalOceanDeleteFail) EditRecord(ctx context.Context, domain string, id int, editRequest *godo.DomainRecordEditRequest) (*godo.DomainRecord, *godo.Response, error) {
	return &godo.DomainRecord{ID: 1}, nil, nil
}

func (m *mockDigitalOceanDeleteFail) Get(ctx context.Context, name string) (*godo.Domain, *godo.Response, error) {
	return &godo.Domain{Name: "example.com"}, nil, nil
}

func (m *mockDigitalOceanDeleteFail) Record(ctx context.Context, domain string, id int) (*godo.DomainRecord, *godo.Response, error) {
	return &godo.DomainRecord{ID: 1}, nil, nil
}

func (m *mockDigitalOceanDeleteFail) Records(ctx context.Context, domain string, opt *godo.ListOptions) ([]godo.DomainRecord, *godo.Response, error) {
	return []godo.DomainRecord{{ID: 1}, {ID: 2}}, nil, nil
}

type mockDigitalOceanRecordsFail struct{}

func (m *mockDigitalOceanRecordsFail) List(context.Context, *godo.ListOptions) ([]godo.Domain, *godo.Response, error) {
	return []godo.Domain{{Name: "foo.com"}, {Name: "bar.com"}}, nil, nil
}

func (m *mockDigitalOceanRecordsFail) Create(context.Context, *godo.DomainCreateRequest) (*godo.Domain, *godo.Response, error) {
	return &godo.Domain{Name: "example.com"}, nil, nil
}

func (m *mockDigitalOceanRecordsFail) CreateRecord(context.Context, string, *godo.DomainRecordEditRequest) (*godo.DomainRecord, *godo.Response, error) {
	return &godo.DomainRecord{ID: 1}, nil, nil
}

func (m *mockDigitalOceanRecordsFail) Delete(context.Context, string) (*godo.Response, error) {
	return nil, nil
}
func (m *mockDigitalOceanRecordsFail) DeleteRecord(ctx context.Context, domain string, id int) (*godo.Response, error) {
	return nil, nil
}
func (m *mockDigitalOceanRecordsFail) EditRecord(ctx context.Context, domain string, id int, editRequest *godo.DomainRecordEditRequest) (*godo.DomainRecord, *godo.Response, error) {
	return &godo.DomainRecord{ID: 1}, nil, nil
}

func (m *mockDigitalOceanRecordsFail) Get(ctx context.Context, name string) (*godo.Domain, *godo.Response, error) {
	return &godo.Domain{Name: "example.com"}, nil, nil
}

func (m *mockDigitalOceanRecordsFail) Record(ctx context.Context, domain string, id int) (*godo.DomainRecord, *godo.Response, error) {
	return nil, nil, fmt.Errorf("Failed to get records")
}

func (m *mockDigitalOceanRecordsFail) Records(ctx context.Context, domain string, opt *godo.ListOptions) ([]godo.DomainRecord, *godo.Response, error) {
	return []godo.DomainRecord{}, nil, fmt.Errorf("Failed to get records")
}

type mockDigitalOceanUpdateRecordsFail struct{}

func (m *mockDigitalOceanUpdateRecordsFail) List(context.Context, *godo.ListOptions) ([]godo.Domain, *godo.Response, error) {
	return []godo.Domain{{Name: "foo.com"}, {Name: "bar.com"}}, nil, nil
}

func (m *mockDigitalOceanUpdateRecordsFail) Create(context.Context, *godo.DomainCreateRequest) (*godo.Domain, *godo.Response, error) {
	return &godo.Domain{Name: "example.com"}, nil, nil
}

func (m *mockDigitalOceanUpdateRecordsFail) CreateRecord(context.Context, string, *godo.DomainRecordEditRequest) (*godo.DomainRecord, *godo.Response, error) {
	return &godo.DomainRecord{ID: 1}, nil, nil
}

func (m *mockDigitalOceanUpdateRecordsFail) Delete(context.Context, string) (*godo.Response, error) {
	return nil, nil
}
func (m *mockDigitalOceanUpdateRecordsFail) DeleteRecord(ctx context.Context, domain string, id int) (*godo.Response, error) {
	return nil, nil
}
func (m *mockDigitalOceanUpdateRecordsFail) EditRecord(ctx context.Context, domain string, id int, editRequest *godo.DomainRecordEditRequest) (*godo.DomainRecord, *godo.Response, error) {
	return nil, nil, fmt.Errorf("Failed to update record")
}

func (m *mockDigitalOceanUpdateRecordsFail) Get(ctx context.Context, name string) (*godo.Domain, *godo.Response, error) {
	return &godo.Domain{Name: "example.com"}, nil, nil
}

func (m *mockDigitalOceanUpdateRecordsFail) Record(ctx context.Context, domain string, id int) (*godo.DomainRecord, *godo.Response, error) {
	return &godo.DomainRecord{ID: 1}, nil, nil
}

func (m *mockDigitalOceanUpdateRecordsFail) Records(ctx context.Context, domain string, opt *godo.ListOptions) ([]godo.DomainRecord, *godo.Response, error) {
	return []godo.DomainRecord{{ID: 1}, {ID: 2}}, nil, nil
}

type mockDigitalOceanDeleteRecordsFail struct{}

func (m *mockDigitalOceanDeleteRecordsFail) List(context.Context, *godo.ListOptions) ([]godo.Domain, *godo.Response, error) {
	return []godo.Domain{{Name: "foo.com"}, {Name: "bar.com"}}, nil, nil
}

func (m *mockDigitalOceanDeleteRecordsFail) Create(context.Context, *godo.DomainCreateRequest) (*godo.Domain, *godo.Response, error) {
	return &godo.Domain{Name: "example.com"}, nil, nil
}

func (m *mockDigitalOceanDeleteRecordsFail) CreateRecord(context.Context, string, *godo.DomainRecordEditRequest) (*godo.DomainRecord, *godo.Response, error) {
	return &godo.DomainRecord{ID: 1}, nil, nil
}

func (m *mockDigitalOceanDeleteRecordsFail) Delete(context.Context, string) (*godo.Response, error) {
	return nil, nil
}
func (m *mockDigitalOceanDeleteRecordsFail) DeleteRecord(ctx context.Context, domain string, id int) (*godo.Response, error) {
	return nil, fmt.Errorf("Failed to delete record")
}
func (m *mockDigitalOceanDeleteRecordsFail) EditRecord(ctx context.Context, domain string, id int, editRequest *godo.DomainRecordEditRequest) (*godo.DomainRecord, *godo.Response, error) {
	return &godo.DomainRecord{ID: 1}, nil, nil
}

func (m *mockDigitalOceanDeleteRecordsFail) Get(ctx context.Context, name string) (*godo.Domain, *godo.Response, error) {
	return &godo.Domain{Name: "example.com"}, nil, nil
}

func (m *mockDigitalOceanDeleteRecordsFail) Record(ctx context.Context, domain string, id int) (*godo.DomainRecord, *godo.Response, error) {
	return &godo.DomainRecord{ID: 1}, nil, nil
}

func (m *mockDigitalOceanDeleteRecordsFail) Records(ctx context.Context, domain string, opt *godo.ListOptions) ([]godo.DomainRecord, *godo.Response, error) {
	return []godo.DomainRecord{{ID: 1, Name: "foobar.ext-dns-test.zalando.to."}, {ID: 2}}, nil, nil
}

type mockDigitalOceanCreateRecordsFail struct{}

func (m *mockDigitalOceanCreateRecordsFail) List(context.Context, *godo.ListOptions) ([]godo.Domain, *godo.Response, error) {
	return []godo.Domain{{Name: "foo.com"}, {Name: "bar.com"}}, nil, nil
}

func (m *mockDigitalOceanCreateRecordsFail) Create(context.Context, *godo.DomainCreateRequest) (*godo.Domain, *godo.Response, error) {
	return &godo.Domain{Name: "example.com"}, nil, nil
}

func (m *mockDigitalOceanCreateRecordsFail) CreateRecord(context.Context, string, *godo.DomainRecordEditRequest) (*godo.DomainRecord, *godo.Response, error) {
	return nil, nil, fmt.Errorf("Failed to create record")
}

func (m *mockDigitalOceanCreateRecordsFail) Delete(context.Context, string) (*godo.Response, error) {
	return nil, nil
}
func (m *mockDigitalOceanCreateRecordsFail) DeleteRecord(ctx context.Context, domain string, id int) (*godo.Response, error) {
	return nil, nil
}
func (m *mockDigitalOceanCreateRecordsFail) EditRecord(ctx context.Context, domain string, id int, editRequest *godo.DomainRecordEditRequest) (*godo.DomainRecord, *godo.Response, error) {
	return &godo.DomainRecord{ID: 1}, nil, nil
}

func (m *mockDigitalOceanCreateRecordsFail) Get(ctx context.Context, name string) (*godo.Domain, *godo.Response, error) {
	return &godo.Domain{Name: "example.com"}, nil, nil
}

func (m *mockDigitalOceanCreateRecordsFail) Record(ctx context.Context, domain string, id int) (*godo.DomainRecord, *godo.Response, error) {
	return &godo.DomainRecord{ID: 1}, nil, nil
}

func (m *mockDigitalOceanCreateRecordsFail) Records(ctx context.Context, domain string, opt *godo.ListOptions) ([]godo.DomainRecord, *godo.Response, error) {
	return []godo.DomainRecord{{ID: 1, Name: "foobar.ext-dns-test.zalando.to."}, {ID: 2}}, nil, nil
}

func TestNewDigitalOceanChanges(t *testing.T) {
	action := DigitalOceanCreate
	endpoints := []*endpoint.Endpoint{{DNSName: "new", Target: "target"}}
	_ = newDigitalOceanChanges(action, endpoints)
}

func TestDigitalOceanZones(t *testing.T) {
	provider := &DigitalOceanProvider{
		Client:       &mockDigitalOceanClient{},
		domainFilter: NewDomainFilter([]string{"com"}),
	}

	zones, err := provider.Zones()
	if err != nil {
		t.Fatal(err)
	}

	validateDigitalOceanZones(t, zones, []godo.Domain{
		{Name: "foo.com"}, {Name: "bar.com"},
	})
}

func TestDigitalOceanRecords(t *testing.T) {
	provider := &DigitalOceanProvider{
		Client: &mockDigitalOceanClient{},
	}

	records, err := provider.Records()
	if err != nil {
		t.Errorf("should not fail, %s", err)
	}

	assert.Equal(t, 3, len(records))

	provider.Client = &mockDigitalOceanRecordsFail{}
	_, err = provider.Records()
	if err == nil {
		t.Errorf("expected to fail, %s", err)
	}
}

func TestDigitalOceanApplyChanges(t *testing.T) {
	changes := &plan.Changes{}
	provider := &DigitalOceanProvider{
		Client: &mockDigitalOceanClient{},
	}
	changes.Create = []*endpoint.Endpoint{{DNSName: "new.ext-dns-test.bar.com", Target: "target"}, {DNSName: "new.ext-dns-test.unexpected.com", Target: "target"}}
	changes.Delete = []*endpoint.Endpoint{{DNSName: "foobar.ext-dns-test.bar.com", Target: "target"}}
	changes.UpdateOld = []*endpoint.Endpoint{{DNSName: "foobar.ext-dns-test.bar.de", Target: "target-old"}}
	changes.UpdateNew = []*endpoint.Endpoint{{DNSName: "foobar.ext-dns-test.foo.com", Target: "target-new"}}
	err := provider.ApplyChanges(changes)
	if err != nil {
		t.Errorf("should not fail, %s", err)
	}
}

func TestNewDigitalOceanProvider(t *testing.T) {
	_ = os.Setenv("DO_TOKEN", "xxxxxxxxxxxxxxxxx")
	_, err := NewDigitalOceanProvider(NewDomainFilter([]string{"ext-dns-test.zalando.to."}), true)
	if err != nil {
		t.Errorf("should not fail, %s", err)
	}
	_ = os.Unsetenv("DO_TOKEN")
	_, err = NewDigitalOceanProvider(NewDomainFilter([]string{"ext-dns-test.zalando.to."}), true)
	if err == nil {
		t.Errorf("expected to fail")
	}
}

func TestDigitalOceanGetRecordID(t *testing.T) {
	p := &DigitalOceanProvider{}
	records := []godo.DomainRecord{
		{
			ID:   1,
			Name: "foo.com",
			Type: "CNAME",
		},
		{
			ID:   2,
			Name: "baz.de",
			Type: "A",
		},
	}
	assert.Equal(t, 1, p.getRecordID(records, godo.DomainRecord{
		Name: "foo.com",
		Type: "CNAME",
	}))

	assert.Equal(t, 0, p.getRecordID(records, godo.DomainRecord{
		Name: "foo.com",
		Type: "A",
	}))
}

func TestDigitalOceanSuitableZone(t *testing.T) {
	domains := []godo.Domain{
		{
			Name: "foo.com",
		},
		{
			Name: "bar.foo.com",
		},
		{
			Name: "baz.com",
		},
	}
	assert.Nil(t, digitalOceanSuitableZone("foo.de", domains))
	assert.Equal(t, "bar.foo.com", digitalOceanSuitableZone("record.bar.foo.com", domains).Name)
}

func validateDigitalOceanZones(t *testing.T, zones []godo.Domain, expected []godo.Domain) {
	require.Len(t, zones, len(expected))

	for i, zone := range zones {
		assert.Equal(t, expected[i].Name, zone.Name)
	}
}
