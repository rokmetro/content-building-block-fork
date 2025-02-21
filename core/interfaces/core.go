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
	"io"

	"github.com/rokwire/core-auth-library-go/v3/tokenauth"
	"go.mongodb.org/mongo-driver/bson"
)

// Services exposes APIs for the driver adapters
type Services interface {
	GetVersion() string
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

	//allApps says if the data is associated with the current app or it is for all the apps within the organization
	GetContentItemsCategories(allApps bool, appID string, orgID string) ([]string, error)
	GetContentItems(allApps bool, appID string, orgID string, ids []string, categoryList []string, offset *int64, limit *int64, order *string) ([]model.ContentItemResponse, error)
	GetContentItem(allApps bool, appID string, orgID string, id string) (*model.ContentItemResponse, error)
	CreateContentItem(allApps bool, appID string, orgID string, category string, data interface{}) (*model.ContentItem, error)
	UpdateContentItem(allApps bool, appID string, orgID string, id string, category string, data interface{}) (*model.ContentItem, error)
	UpdateContentItemData(allApps bool, appID string, orgID string, id string, category string, data interface{}) (*model.ContentItem, error)
	DeleteContentItem(allApps bool, appID string, orgID string, id string) error
	DeleteContentItemByCategory(allApps bool, appID string, orgID string, id string, category string) error

	UploadImage(imageBytes []byte, path string, spec model.ImageSpec) (*string, error)
	GetProfileImage(userID string, imageType string) ([]byte, error)
	UploadProfileImage(userID string, bytes []byte) error
	DeleteProfileImage(userID string) error

	UploadVoiceRecord(userID string, bytes []byte) error
	GetVoiceRecord(userID string) ([]byte, error)
	DeleteVoiceRecord(userID string) error

	GetTwitterPosts(userID string, twitterQueryParams string, force bool) (map[string]interface{}, error)

	CreateDataContentItem(claims *tokenauth.Claims, item *model.DataContentItem) (*model.DataContentItem, error)
	GetDataContentItem(claims *tokenauth.Claims, key string) (*model.DataContentItem, error)
	UpdateDataContentItem(claims *tokenauth.Claims, item *model.DataContentItem) (*model.DataContentItem, error)
	DeleteDataContentItem(claims *tokenauth.Claims, key string) error
	GetDataContentItems(claims *tokenauth.Claims, category string) ([]*model.DataContentItem, error)

	CreateCategory(claims *tokenauth.Claims, item *model.Category) (*model.Category, error)
	GetCategory(claims *tokenauth.Claims, name string) (*model.Category, error)
	UpdateCategory(claims *tokenauth.Claims, item *model.Category) (*model.Category, error)
	DeleteCategory(claims *tokenauth.Claims, name string) error

	UploadFileContentItem(file io.Reader, claims *tokenauth.Claims, fileName string, category string) error
	GetFileContentItem(claims *tokenauth.Claims, fileName string, category string) (io.ReadCloser, error)
	GetFileContentUploadURLs(claims *tokenauth.Claims, fileNames []string, entityID string, category string) ([]model.FileContentItemRef, error)
	GetFileContentDownloadURLs(claims *tokenauth.Claims, fileKeys []string, entityID string, category string) ([]model.FileContentItemRef, error)
	DeleteFileContentItem(claims *tokenauth.Claims, fileName string, category string) error
}
