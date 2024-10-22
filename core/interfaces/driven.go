// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package interfaces

import (
	"content/core/model"

	"go.mongodb.org/mongo-driver/bson"
)

// Storage is used by core to storage data - DB storage adapter, file storage adapter etc
type Storage interface {
	PerformTransaction(transaction func(storage Storage) error) error

	GetStudentGuides(appID string, orgID string, ids []string) ([]bson.M, error)
	GetStudentGuide(appID string, orgID string, id string) (bson.M, error)
	CreateStudentGuide(appID string, orgID string, item bson.M) (bson.M, error)
	UpdateStudentGuide(appID string, orgID string, id string, item bson.M) (bson.M, error)
	DeleteStudentGuide(appID string, orgID string, id string) error

	GetHealthLocations(appID string, orgID string, ids []string) ([]bson.M, error)
	GetHealthLocation(appID string, orgID string, id string) (bson.M, error)
	CreateHealthLocation(appID string, orgID string, item bson.M) (bson.M, error)
	UpdateHealthLocation(appID string, orgID string, id string, item bson.M) (bson.M, error)
	DeleteHealthLocation(appID string, orgID string, id string) error

	GetContentItemsCategories(appID *string, orgID string) ([]string, error)
	FindContentItems(appID *string, orgID string, ids []string, categoryList []string, offset *int64, limit *int64, order *string) ([]model.ContentItem, error)
	GetContentItems(appID *string, orgID string, ids []string, categoryList []string, offset *int64, limit *int64, order *string) ([]model.ContentItemResponse, error)
	GetContentItem(appID *string, orgID string, id string) (*model.ContentItemResponse, error)
	CreateContentItem(item model.ContentItem) (*model.ContentItem, error)
	UpdateContentItem(appID *string, orgID string, id string, category string, data interface{}) (*model.ContentItem, error)
	DeleteContentItem(appID *string, orgID string, id string) error
	SaveContentItem(item model.ContentItem) error

	//Used for multi-tenancy for already exisiting data.
	//To be removed when this is applied to all environments.
	FindAllContentItems() ([]model.ContentItemResponse, error)
	StoreMultiTenancyData(appID string, orgID string) error
	///

	CreateDataContentItem(item *model.DataContentItem) (*model.DataContentItem, error)
	FindDataContentItem(appID *string, orgID string, key string) (*model.DataContentItem, error)
	UpdateDataContentItem(appID *string, orgID string, item *model.DataContentItem) (*model.DataContentItem, error)
	DeleteDataContentItem(appID *string, orgID string, key string) error
	FindDataContentItems(appID *string, orgID string, key string) ([]*model.DataContentItem, error)

	CreateCategory(item *model.Category) (*model.Category, error)
	FindCategory(appID *string, orgID string, name string) (*model.Category, error)
	UpdateCategory(appID *string, orgID string, item *model.Category) (*model.Category, error)
	DeleteCategory(appID *string, orgID string, key string) error
}

// Core BB interface
type Core interface {
	LoadDeletedMemberships() ([]model.DeletedUserData, error)
}
