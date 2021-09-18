package mangad

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"mime"
	"os"
	"path"
	"sort"
	"strconv"
	"strings"

	"github.com/clementd64/mangad/pkg/schema/tachidesk"
	"github.com/clementd64/mangad/pkg/schema/tachiyomi"
)

func getChapterFilename(c tachidesk.Chapter) string {
	return fmt.Sprintf("%s.%d.zip", base64.RawURLEncoding.EncodeToString([]byte(c.Name)), c.Index)
}

type Manga struct {
	*tachidesk.Manga
	*Mangad
	Title string
}

func (m *Manga) GetChapters() ([]tachidesk.Chapter, error) {
	chapters, err := m.getChapters(m.Id)
	if err != nil {
		return nil, err
	}

	sort.Slice(chapters, func(i, j int) bool { return chapters[i].Index < chapters[j].Index })

	return chapters, nil
}

func (m *Manga) DownloadChapter(c tachidesk.Chapter) error {
	return m.downloadChapter(c.Index, path.Join(m.Title, getChapterFilename(c)))
}

func (m *Manga) downloadChapter(index int, filename string) error {
	chapter, err := m.getChapter(m.Id, index)
	if err != nil {
		return err
	}

	dir, err := ioutil.TempDir("", "manga-")
	if err != nil {
		return err
	}
	defer os.RemoveAll(dir)

	for i := 0; i < chapter.PageCount; i++ {
		image, mimetype, err := m.getPage(m.Id, chapter.Index, i)
		if err != nil {
			return err
		}

		exts, err := mime.ExtensionsByType(mimetype)
		if err != nil {
			return err
		}

		ext := ".bin"
		if exts != nil {
			ext = exts[len(exts)-1]
		}

		file := path.Join(dir, fmt.Sprintf("%06d", i)+ext)
		err = ioutil.WriteFile(file, image, 0644)
		if err != nil {
			return err
		}
	}

	if _, err := os.Stat(m.Title); os.IsNotExist(err) {
		if err := os.MkdirAll(m.Title, 0775); err != nil {
			return err
		}
	}

	return zipFolder(dir, filename)
}

func (m *Manga) GetDownloaded() (map[int]bool, error) {
	downloaded := map[int]bool{}

	if _, err := os.Stat(m.Title); os.IsNotExist(err) {
		return downloaded, nil
	}

	files, err := os.ReadDir(m.Title)
	if err != nil {
		return nil, err
	}

	for _, file := range files {
		part := strings.Split(file.Name(), ".")

		if len(part) < 2 {
			continue
		}

		num, err := strconv.Atoi(part[len(part)-2])
		if err != nil {
			continue
		}

		downloaded[num] = true
	}

	return downloaded, nil
}

func (m *Manga) Process() error {
	if err := m.exportDetails(); err != nil {
		return err
	}
	if err := m.exportThumbnail(); err != nil {
		return err
	}

	chapters, err := m.GetChapters()
	if err != nil {
		return err
	}

	downloaded, err := m.GetDownloaded()
	if err != nil {
		return err
	}

	for _, chapter := range chapters {
		if _, ok := downloaded[chapter.Index]; ok {
			continue
		}

		if err := m.DownloadChapter(chapter); err != nil {
			return err
		}
	}

	return nil
}

// https://tachiyomi.org/help/guides/local-manga/#editing-local-manga-details
func (m *Manga) exportDetails() error {
	status := "0"
	switch m.Status {
	case "ONGOING":
		status = "1"
	case "COMPLETED":
		status = "2"
	case "LICENSED":
		status = "3"
	}

	details := tachiyomi.LocalDetail{
		Title:       m.Title,
		Author:      m.Author,
		Artist:      m.Artist,
		Description: m.Description,
		Genre:       m.Genre,
		Status:      status,
		StatusValues: []string{
			"0 = Unknown",
			"1 = Ongoing",
			"2 = Completed",
			"3 = Licensed",
		},
	}

	file, err := json.Marshal(details)
	if err != nil {
		return err
	}

	return os.WriteFile(path.Join(m.Title, "details.json"), file, 0664)
}

// https://tachiyomi.org/help/guides/local-manga/#using-a-custom-cover-image
func (m *Manga) exportThumbnail() error {
	if _, err := os.Stat(path.Join(m.Title, "cover.jpg")); os.IsExist(err) {
		return nil
	}

	body, _, err := m.getThumbnail(m.Id)
	if err != nil {
		return err
	}

	return os.WriteFile(path.Join(m.Title, "cover.jpg"), body, 0664)
}
