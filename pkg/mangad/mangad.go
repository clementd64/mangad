package mangad

import (
	"bytes"
	"compress/gzip"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"strconv"
	"sync"

	"github.com/clementd64/mangad/pkg/schema/backup"
	"github.com/clementd64/mangad/pkg/schema/tachidesk"
	"github.com/h2non/filetype"
	"google.golang.org/protobuf/proto"
)

type Mangad struct {
	BaseURL string
	mutex   sync.Mutex
}

func New(baseURL string) *Mangad {
	return &Mangad{
		BaseURL: baseURL + "api/",
	}
}

func (t *Mangad) Ping() error {
	res, err := http.Get(t.BaseURL + "v1/settings/about")
	if err != nil {
		return err
	}
	res.Body.Close()
	return nil
}

func (t *Mangad) fetch(endpoint string, data interface{}) error {
	log.Print(endpoint)
	res, err := http.Get(t.BaseURL + endpoint)
	if err != nil {
		return err
	}
	defer res.Body.Close()

	body, err := io.ReadAll(res.Body)
	if err != nil {
		return err
	}

	if res.StatusCode >= 300 {
		return errors.New("failed to fetch " + endpoint + " : " + string(body))
	}

	if data != nil {
		if err := json.Unmarshal(body, data); err != nil {
			return err
		}
	}

	return nil
}

func (t *Mangad) sourceExist(sourceId int64) (bool, error) {
	source := &tachidesk.Source{}
	if err := t.fetch(fmt.Sprintf("v1/source/%d", sourceId), source); err != nil {
		return false, err
	}
	return source.Name != "", nil
}

func (t *Mangad) updateExtensionList() error {
	t.mutex.Lock()
	err := t.fetch("v1/extension/list", nil)
	t.mutex.Unlock()
	return err
}

func (t *Mangad) installExtension(pkgName string) error {
	t.mutex.Lock()
	err := t.fetch("v1/extension/install/"+pkgName, nil)
	t.mutex.Unlock()
	return err
}

func (t *Mangad) findManga(sourceId int64, url string) (*tachidesk.Manga, error) {
	list := []tachidesk.Manga{}
	if err := t.fetch("v1/category/0", &list); err != nil {
		return nil, err
	}

	for _, manga := range list {
		if manga.SourceId == strconv.FormatInt(sourceId, 10) && manga.Url == url {
			return &manga, nil
		}
	}

	return nil, nil
}

func (t *Mangad) addManga(sourceId int64, url string) error {
	backup := &backup.Backup{
		BackupManga: []*backup.BackupManga{
			{
				Source: &sourceId,
				Url:    &url,
			},
		},
	}

	out, err := proto.Marshal(backup)
	if err != nil {
		return nil
	}

	var b bytes.Buffer
	gz := gzip.NewWriter(&b)
	if _, err := gz.Write(out); err != nil {
		return err
	}
	if err := gz.Flush(); err != nil {
		return err
	}
	if err := gz.Close(); err != nil {
		return err
	}

	log.Printf("v1/backup/import - %d - %s", sourceId, url)
	t.mutex.Lock()
	res, err := http.Post(t.BaseURL+"v1/backup/import", "", &b)
	t.mutex.Unlock()
	if err != nil {
		return err
	}

	if res.StatusCode >= 300 {
		return errors.New("failed to restore backup")
	}

	return nil
}

func (t *Mangad) getChapters(mangaId int) ([]tachidesk.Chapter, error) {
	chapters := []tachidesk.Chapter{}
	if err := t.fetch(fmt.Sprintf("v1/manga/%d/chapters?onlineFetch=true", mangaId), &chapters); err != nil {
		return nil, err
	}
	return chapters, nil
}

func (t *Mangad) getChapter(mangaId, chapterId int) (*tachidesk.Chapter, error) {
	chapter := &tachidesk.Chapter{}
	if err := t.fetch(fmt.Sprintf("v1/manga/%d/chapter/%d", mangaId, chapterId), chapter); err != nil {
		return nil, err
	}
	return chapter, nil
}

func (t *Mangad) getPage(mangaId, chapterId, page int) ([]byte, string, error) {
	url := fmt.Sprintf("v1/manga/%d/chapter/%d/page/%d", mangaId, chapterId, page)
	return t.downloadImage(t.BaseURL + url)
}

func (t *Mangad) getThumbnail(mangaId int) ([]byte, string, error) {
	return t.downloadImage(t.BaseURL + fmt.Sprintf("v1/manga/%d/thumbnail", mangaId))
}

func (t *Mangad) downloadImage(url string) ([]byte, string, error) {
	res, err := http.Get(url)
	if err != nil {
		return nil, "", err
	}
	defer res.Body.Close()

	body, err := io.ReadAll(res.Body)
	if err != nil {
		return nil, "", err
	}

	// image type may be broken, so we check it using magic number
	kind, err := filetype.Match(body)
	if err != nil {
		return nil, "", err
	}

	if kind != filetype.Unknown {
		return body, kind.MIME.Value, nil
	}

	return body, res.Header.Get("Content-Type"), nil
}

func (t *Mangad) Manga(source int64, pkgName, url, title string) (*Manga, error) {
	ok, err := t.sourceExist(source)
	if err != nil {
		return nil, err
	}
	if !ok {
		if err := t.updateExtensionList(); err != nil {
			return nil, err
		}
		if err := t.installExtension(pkgName); err != nil {
			return nil, err
		}
	}

	mangaDetail, err := t.findManga(source, url)
	if err != nil {
		return nil, err
	}
	if mangaDetail == nil {
		if err := t.addManga(source, url); err != nil {
			return nil, err
		}
		if mangaDetail, err = t.findManga(source, url); err != nil {
			return nil, err
		}
	}

	manga := &Manga{
		Manga:  mangaDetail,
		Mangad: t,
		Title:  title,
	}

	return manga, nil
}

func (t *Mangad) Run(conf map[int64][]Config) {
	var wg sync.WaitGroup
	for _, c := range conf {
		wg.Add(1)
		go func(c []Config) {
			defer wg.Done()
			t.processList(c)
		}(c)
	}

	wg.Wait()
}

func (t *Mangad) processList(conf []Config) {
	for _, c := range conf {
		m, err := t.Manga(c.Source, c.PkgName, c.Url, c.Title)
		if err != nil {
			log.Print(err)
			continue
		}

		if err := m.Process(); err != nil {
			log.Print(err)
		}
	}
}
