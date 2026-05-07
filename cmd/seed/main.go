// cmd/seed/main.go
package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"strings"

	"NetCentricLab-MangaHub/pkg/database"
)

// MangaDexResponse maps the API structure including authors and cover_art
type MangaDexResponse struct {
	Data []struct {
		ID         string `json:"id"`
		Attributes struct {
			Title       map[string]string `json:"title"`
			Description map[string]string `json:"description"`
			Status      string            `json:"status"`
		} `json:"attributes"`
		Relationships []struct {
			Type       string `json:"type"`
			Attributes struct {
				Name     string `json:"name"`     // For authors
				FileName string `json:"fileName"` // For cover_art
			} `json:"attributes"`
		} `json:"relationships"`
	} `json:"data"`
}

type mangaEntry struct {
	ID            string
	Title         string
	Author        string
	Genres        []string
	Status        string
	TotalChapters int
	Description   string
	CoverURL      string
}

// Global map to track duplicates and prevent overlap
var existingTitles = make(map[string]bool)

// Helper to fetch the real cover image for manually hardcoded manga
func fetchCoverForManualSeed(title string) string {
	encodedTitle := url.QueryEscape(title)
	apiURL := fmt.Sprintf("https://api.mangadex.org/manga?title=%s&includes[]=cover_art&limit=1", encodedTitle)

	resp, err := http.Get(apiURL)
	if err != nil || resp.StatusCode != 200 {
		return "https://example.com/covers/default.jpg"
	}
	defer resp.Body.Close()

	var dexResp MangaDexResponse
	// Fix applied: Using  to access the first item in the Data array
	if err := json.NewDecoder(resp.Body).Decode(&dexResp); err == nil && len(dexResp.Data) > 0 {
		mangaID := dexResp.Data[0].ID
		for _, rel := range dexResp.Data[0].Relationships {
			if rel.Type == "cover_art" && rel.Attributes.FileName != "" {
				return fmt.Sprintf("https://uploads.mangadex.org/covers/%s/%s", mangaID, rel.Attributes.FileName)
			}
		}
	}
	return "https://example.com/covers/default.jpg" // Fallback
}

// Function to inject hardcoded manga entries
func seedManual(db *sql.DB) {
	entries := []mangaEntry{
		// --- SHOUNEN (25) ---
		{ID: "one-piece", Title: "One Piece", Author: "Oda Eiichiro", Genres: []string{"Action", "Adventure", "Shounen"}, Status: "ongoing", TotalChapters: 1115, Description: "A young pirate's adventure to find the legendary treasure One Piece."},
		{ID: "naruto", Title: "Naruto", Author: "Kishimoto Masashi", Genres: []string{"Action", "Adventure", "Shounen"}, Status: "completed", TotalChapters: 700, Description: "A young ninja's journey to become the greatest Hokage."},
		{ID: "demon-slayer", Title: "Demon Slayer", Author: "Gotouge Koyoharu", Genres: []string{"Action", "Shounen"}, Status: "completed", TotalChapters: 205, Description: "A boy becomes a demon slayer to cure his demonified sister."},
		{ID: "my-hero-academia", Title: "My Hero Academia", Author: "Horikoshi Kouhei", Genres: []string{"Action", "Shounen"}, Status: "completed", TotalChapters: 430, Description: "A boy born without powers in a superhero-filled world."},
		{ID: "jujutsu-kaisen", Title: "Jujutsu Kaisen", Author: "Akutami Gege", Genres: []string{"Action", "Supernatural", "Shounen"}, Status: "ongoing", TotalChapters: 260, Description: "A high schooler swallows a cursed finger and enters a world of sorcery."},
		{ID: "bleach", Title: "Bleach", Author: "Kubo Tite", Genres: []string{"Action", "Shounen"}, Status: "completed", TotalChapters: 686, Description: "Ichigo Kurosaki gains the powers of a Soul Reaper."},
		{ID: "hunter-x-hunter", Title: "Hunter x Hunter", Author: "Togashi Yoshihiro", Genres: []string{"Adventure", "Fantasy", "Shounen"}, Status: "ongoing", TotalChapters: 400, Description: "Gon Freecss seeks to find his father by becoming a Hunter."},
		{ID: "chainsaw-man", Title: "Chainsaw Man", Author: "Fujimoto Tatsuki", Genres: []string{"Action", "Horror", "Shounen"}, Status: "ongoing", TotalChapters: 165, Description: "A debt-ridden boy merges with his chainsaw dog to hunt devils."},
		{ID: "black-clover", Title: "Black Clover", Author: "Tabata Yuuki", Genres: []string{"Fantasy", "Action", "Shounen"}, Status: "ongoing", TotalChapters: 370, Description: "Asta strives to become the Wizard King without any magic power."},
		{ID: "fullmetal-alchemist", Title: "Fullmetal Alchemist", Author: "Arakawa Hiromu", Genres: []string{"Adventure", "Steampunk", "Shounen"}, Status: "completed", TotalChapters: 108, Description: "Two brothers use alchemy to search for the Philosopher's Stone."},
		{ID: "dragon-ball", Title: "Dragon Ball", Author: "Toriyama Akira", Genres: []string{"Adventure", "Martial Arts", "Shounen"}, Status: "completed", TotalChapters: 519, Description: "Son Goku explores the world in search of the Dragon Balls."},
		{ID: "spy-x-family", Title: "Spy x Family", Author: "Endo Tatsuya", Genres: []string{"Comedy", "Action", "Shounen"}, Status: "ongoing", TotalChapters: 100, Description: "A spy, an assassin, and a telepath form a fake family."},
		{ID: "haikyuu", Title: "Haikyuu!!", Author: "Furudate Haruichi", Genres: []string{"Sports", "Shounen"}, Status: "completed", TotalChapters: 402, Description: "A short boy aims to conquer the world of high school volleyball."},
		{ID: "death-note", Title: "Death Note", Author: "Ohba Tsugumi", Genres: []string{"Psychological", "Supernatural", "Shounen"}, Status: "completed", TotalChapters: 108, Description: "A student finds a notebook that kills anyone whose name is written in it."},
		{ID: "gintama", Title: "Gintama", Author: "Sorachi Hideaki", Genres: []string{"Comedy", "Samurai", "Shounen"}, Status: "completed", TotalChapters: 704, Description: "An eccentric samurai works as a freelancer in an alien-occupied Edo."},
		{ID: "blue-lock", Title: "Blue Lock", Author: "Kaneshiro Muneyuki", Genres: []string{"Sports", "Shounen"}, Status: "ongoing", TotalChapters: 260, Description: "300 strikers compete in a prison-like facility to become Japan's best."},
		{ID: "dr-stone", Title: "Dr. Stone", Author: "Inagaki Riichiro", Genres: []string{"Sci-Fi", "Adventure", "Shounen"}, Status: "completed", TotalChapters: 232, Description: "A scientific genius wakes up thousands of years after humanity was petrified."},
		{ID: "fire-force", Title: "Fire Force", Author: "Ohkubo Atsushi", Genres: []string{"Action", "Sci-Fi", "Shounen"}, Status: "completed", TotalChapters: 304, Description: "Special fire brigades fight spontaneous human combustion."},
		{ID: "the-promised-neverland", Title: "The Promised Neverland", Author: "Shirai Kaiu", Genres: []string{"Thriller", "Mystery", "Shounen"}, Status: "completed", TotalChapters: 181, Description: "Orphans discover the dark secret behind their idyllic home."},
		{ID: "slam-dunk", Title: "Slam Dunk", Author: "Inoue Takehiko", Genres: []string{"Sports", "Shounen"}, Status: "completed", TotalChapters: 276, Description: "A delinquent joins the basketball team to impress a girl."},
		{ID: "fairy-tail", Title: "Fairy Tail", Author: "Mashima Hiro", Genres: []string{"Fantasy", "Action", "Shounen"}, Status: "completed", TotalChapters: 545, Description: "A wizard joins the rowdy Fairy Tail guild for various missions."},
		{ID: "blue-exorcist", Title: "Blue Exorcist", Author: "Kato Kazue", Genres: []string{"Supernatural", "Action", "Shounen"}, Status: "ongoing", TotalChapters: 150, Description: "The son of Satan decides to become an exorcist to fight his father."},
		{ID: "sakamoto-days", Title: "Sakamoto Days", Author: "Suzuki Yuto", Genres: []string{"Action", "Comedy", "Shounen"}, Status: "ongoing", TotalChapters: 170, Description: "A legendary retired hitman tries to live a peaceful life as a store owner."},
		{ID: "fist-of-the-north-star", Title: "Fist of the North Star", Author: "Hara Tetsuo", Genres: []string{"Action", "Martial Arts", "Shounen"}, Status: "completed", TotalChapters: 245, Description: "Kenshiro wanders a post-apocalyptic wasteland using a lethal martial art."},
		{ID: "assassination-classroom", Title: "Assassination Classroom", Author: "Matsui Yusei", Genres: []string{"Comedy", "Action", "Shounen"}, Status: "completed", TotalChapters: 180, Description: "Students must kill their alien teacher before he destroys Earth."},

		// --- SEINEN (25) ---
		{ID: "berserk", Title: "Berserk", Author: "Miura Kentaro", Genres: []string{"Dark Fantasy", "Action", "Seinen"}, Status: "ongoing", TotalChapters: 375, Description: "The story of Guts, a lone mercenary seeking revenge against his former friend."},
		{ID: "vinland-saga", Title: "Vinland Saga", Author: "Yukimura Makoto", Genres: []string{"Historical", "Action", "Seinen"}, Status: "ongoing", TotalChapters: 210, Description: "Thorfinn seeks revenge for his father's death while discovering the true meaning of peace."},
		{ID: "vagabond", Title: "Vagabond", Author: "Inoue Takehiko", Genres: []string{"Historical", "Samurai", "Seinen"}, Status: "hiatus", TotalChapters: 327, Description: "A fictionalized account of the life of the legendary swordsman Musashi Miyamoto."},
		{ID: "monster", Title: "Monster", Author: "Urasawa Naoki", Genres: []string{"Mystery", "Thriller", "Seinen"}, Status: "completed", TotalChapters: 162, Description: "A surgeon's life spirals into chaos after he saves a boy who becomes a serial killer."},
		{ID: "tokyo-ghoul", Title: "Tokyo Ghoul", Author: "Ishida Sui", Genres: []string{"Horror", "Supernatural", "Seinen"}, Status: "completed", TotalChapters: 143, Description: "A student becomes a half-ghoul after a chance encounter and must navigate ghoul society."},
		{ID: "kingdom", Title: "Kingdom", Author: "Hara Yasuhisa", Genres: []string{"Historical", "Military", "Seinen"}, Status: "ongoing", TotalChapters: 800, Description: "A war orphan aims to become the greatest general in ancient China."},
		{ID: "oyasumi-punpun", Title: "Goodnight Punpun", Author: "Asano Inio", Genres: []string{"Drama", "Psychological", "Seinen"}, Status: "completed", TotalChapters: 147, Description: "The coming-of-age story of a boy depicted as a simple bird-like drawing."},
		{ID: "20th-century-boys", Title: "20th Century Boys", Author: "Urasawa Naoki", Genres: []string{"Sci-Fi", "Mystery", "Seinen"}, Status: "completed", TotalChapters: 249, Description: "Friends discover a cult leader's plan to destroy the world based on their childhood games."},
		{ID: "kaguya-sama", Title: "Kaguya-sama: Love is War", Author: "Akasaka Aka", Genres: []string{"Comedy", "Romance", "Seinen"}, Status: "completed", TotalChapters: 281, Description: "Two geniuses try to trick the other into confessing their love first."},
		{ID: "grand-blue", Title: "Grand Blue Dreaming", Author: "Inoue Kenji", Genres: []string{"Comedy", "Slice of Life", "Seinen"}, Status: "ongoing", TotalChapters: 93, Description: "A college student joins a diving club full of heavy drinkers and chaos."},
		{ID: "gantz", Title: "Gantz", Author: "Oku Hiroya", Genres: []string{"Sci-Fi", "Action", "Seinen"}, Status: "completed", TotalChapters: 383, Description: "Dead people are forced to play a survival game against aliens in Tokyo."},
		{ID: "one-punch-man", Title: "One Punch Man", Author: "ONE", Genres: []string{"Action", "Comedy", "Seinen"}, Status: "ongoing", TotalChapters: 200, Description: "Saitama can defeat any foe with a single punch but suffers from boredom."},
		{ID: "made-in-abyss", Title: "Made in Abyss", Author: "Tsukushi Akihito", Genres: []string{"Adventure", "Fantasy", "Seinen"}, Status: "ongoing", TotalChapters: 67, Description: "A girl descends into a giant, mysterious hole in the earth searching for her mother."},
		{ID: "golden-kamuy", Title: "Golden Kamuy", Author: "Noda Satoru", Genres: []string{"Historical", "Adventure", "Seinen"}, Status: "completed", TotalChapters: 314, Description: "A veteran and an Ainu girl hunt for hidden gold in Hokkaido."},
		{ID: "mushishi", Title: "Mushishi", Author: "Urushibara Yuki", Genres: []string{"Supernatural", "Slice of Life", "Seinen"}, Status: "completed", TotalChapters: 50, Description: "Ginko travels to research primitive lifeforms known as Mushi."},
		{ID: "bungo-stray-dogs", Title: "Bungo Stray Dogs", Author: "Asagiri Kafka", Genres: []string{"Action", "Supernatural", "Seinen"}, Status: "ongoing", TotalChapters: 115, Description: "Gifted individuals with powers named after literary figures solve crimes."},
		{ID: "blue-period", Title: "Blue Period", Author: "Yamaguchi Tsubasa", Genres: []string{"Drama", "Art", "Seinen"}, Status: "ongoing", TotalChapters: 65, Description: "A popular student discovers a passion for painting and aims for art school."},
		{ID: "real", Title: "Real", Author: "Inoue Takehiko", Genres: []string{"Drama", "Sports", "Seinen"}, Status: "ongoing", TotalChapters: 96, Description: "The lives of three young men connected through wheelchair basketball."},
		{ID: "dorohedoro", Title: "Dorohedoro", Author: "Hayashida Q", Genres: []string{"Dark Fantasy", "Comedy", "Seinen"}, Status: "completed", TotalChapters: 167, Description: "A lizard-headed man hunts sorcerers to find the one who cursed him."},
		{ID: "land-of-the-lustrous", Title: "Land of the Lustrous", Author: "Ichikawa Haruko", Genres: []string{"Fantasy", "Sci-Fi", "Seinen"}, Status: "completed", TotalChapters: 108, Description: "In a future of immortal gemstone people, Phos seeks a role in society."},
		{ID: "hellsing", Title: "Hellsing", Author: "Hirano Kouta", Genres: []string{"Action", "Supernatural", "Seinen"}, Status: "completed", TotalChapters: 95, Description: "The vampire Alucard protects England from supernatural threats."},
		{ID: "parasyte", Title: "Parasyte", Author: "Iwaaki Hitoshi", Genres: []string{"Horror", "Sci-Fi", "Seinen"}, Status: "completed", TotalChapters: 64, Description: "A boy coexists with an alien parasite that replaced his hand."},
		{ID: "the-fable", Title: "The Fable", Author: "Minami Katsuhisa", Genres: []string{"Action", "Comedy", "Seinen"}, Status: "completed", TotalChapters: 240, Description: "A genius hitman is ordered to live as a normal person for one year."},
		{ID: "dungeon-meshi", Title: "Dungeon Meshi", Author: "Kui Ryoko", Genres: []string{"Fantasy", "Comedy", "Seinen"}, Status: "completed", TotalChapters: 97, Description: "Adventurers survive in a dungeon by eating the monsters they kill."},
		{ID: "dead-dead-demons", Title: "Dead Dead Demon's Dededede Destruction", Author: "Asano Inio", Genres: []string{"Sci-Fi", "Slice of Life", "Seinen"}, Status: "completed", TotalChapters: 100, Description: "High school girls live their daily lives while an alien ship hangs over Tokyo."},

		// --- SHOUJO (25) ---
		{ID: "fruits-basket", Title: "Fruits Basket", Author: "Takaya Natsuki", Genres: []string{"Romance", "Drama", "Shoujo"}, Status: "completed", TotalChapters: 136, Description: "Tohru Honda lives with the Sohma family, who turn into zodiac animals when hugged."},
		{ID: "sailor-moon", Title: "Sailor Moon", Author: "Takeuchi Naoko", Genres: []string{"Magical Girl", "Fantasy", "Shoujo"}, Status: "completed", TotalChapters: 60, Description: "Usagi Tsukino leads a team of Sailor Guardians to protect the Earth."},
		{ID: "nana", Title: "Nana", Author: "Yazawa Ai", Genres: []string{"Drama", "Music", "Shoujo"}, Status: "hiatus", TotalChapters: 84, Description: "Two girls named Nana meet on a train and share a flat in Tokyo."},
		{ID: "ouran-host-club", Title: "Ouran High School Host Club", Author: "Hatori Bisco", Genres: []string{"Comedy", "Romance", "Shoujo"}, Status: "completed", TotalChapters: 83, Description: "A girl accidentally breaks an expensive vase and joins a club of handsome hosts."},
		{ID: "yona-of-the-dawn", Title: "Yona of the Dawn", Author: "Kusanagi Mizuho", Genres: []string{"Fantasy", "Adventure", "Shoujo"}, Status: "ongoing", TotalChapters: 255, Description: "A princess flees her castle and seeks the legendary Four Dragons to reclaim her kingdom."},
		{ID: "kamisama-kiss", Title: "Kamisama Kiss", Author: "Suzuki Julietta", Genres: []string{"Supernatural", "Romance", "Shoujo"}, Status: "completed", TotalChapters: 149, Description: "A homeless girl becomes the local Earth Deity and falls for her fox familiar."},
		{ID: "ao-haru-ride", Title: "Ao Haru Ride", Author: "Sakisaka Io", Genres: []string{"Romance", "Drama", "Shoujo"}, Status: "completed", TotalChapters: 49, Description: "Futaba reunites with her middle school crush, but he has changed completely."},
		{ID: "kimi-ni-todoke", Title: "Kimi ni Todoke", Author: "Shiina Karuho", Genres: []string{"Romance", "Slice of Life", "Shoujo"}, Status: "completed", TotalChapters: 123, Description: "A misunderstood girl starts to open up thanks to a popular classmate."},
		{ID: "skip-beat", Title: "Skip Beat!", Author: "Nakamura Yoshiki", Genres: []string{"Drama", "Comedy", "Shoujo"}, Status: "ongoing", TotalChapters: 310, Description: "Kyoko enters show business to get revenge on her ex-boyfriend."},
		{ID: "cardcaptor-sakura", Title: "Cardcaptor Sakura", Author: "CLAMP", Genres: []string{"Magical Girl", "Adventure", "Shoujo"}, Status: "completed", TotalChapters: 50, Description: "Sakura must retrieve a deck of magical cards she accidentally scattered."},
		{ID: "maid-sama", Title: "Maid-sama!", Author: "Fujiwara Hiro", Genres: []string{"Comedy", "Romance", "Shoujo"}, Status: "completed", TotalChapters: 85, Description: "The student council president secretly works at a maid cafe."},
		{ID: "banana-fish", Title: "Banana Fish", Author: "Yoshida Akimi", Genres: []string{"Action", "Thriller", "Shoujo"}, Status: "completed", TotalChapters: 110, Description: "A street gang leader investigates a mysterious drug in 1980s New York."},
		{ID: "itazura-na-kiss", Title: "Itazura na Kiss", Author: "Tada Kaoru", Genres: []string{"Romance", "Comedy", "Shoujo"}, Status: "completed", TotalChapters: 23, Description: "A clumsy girl lives with her genius crush after her house is destroyed."},
		{ID: "say-i-love-you", Title: "Say I Love You", Author: "Hazuki Kanae", Genres: []string{"Romance", "Drama", "Shoujo"}, Status: "completed", TotalChapters: 72, Description: "A quiet girl befriends the most popular boy in school."},
		{ID: "orange", Title: "Orange", Author: "Takano Ichigo", Genres: []string{"Drama", "Sci-Fi", "Shoujo"}, Status: "completed", TotalChapters: 38, Description: "A girl receives a letter from her future self to save a friend."},
		{ID: "vampire-knight", Title: "Vampire Knight", Author: "Hino Matsuri", Genres: []string{"Supernatural", "Romance", "Shoujo"}, Status: "completed", TotalChapters: 93, Description: "A school divides its students into a Day Class and a Night Class of vampires."},
		{ID: "lovely-complex", Title: "Lovely Complex", Author: "Nakahara Aya", Genres: []string{"Comedy", "Romance", "Shoujo"}, Status: "completed", TotalChapters: 68, Description: "A tall girl and a short boy navigate their height-defying romance."},
		{ID: "basara", Title: "Basara", Author: "Tamura Yumi", Genres: []string{"Fantasy", "Adventure", "Shoujo"}, Status: "completed", TotalChapters: 107, Description: "A girl takes her brother's place to lead a revolution in a post-apocalyptic Japan."},
		{ID: "rose-of-versailles", Title: "The Rose of Versailles", Author: "Ikeda Riyoko", Genres: []string{"Historical", "Drama", "Shoujo"}, Status: "completed", TotalChapters: 82, Description: "Oscar François de Jarjayes lives as a man to guard Marie Antoinette."},
		{ID: "glass-mask", Title: "Glass Mask", Author: "Miuchi Suzue", Genres: []string{"Drama", "Arts", "Shoujo"}, Status: "ongoing", TotalChapters: 49, Description: "Two rival actresses compete for the legendary role of the Crimson Goddess."},
		{ID: "special-a", Title: "Special A", Author: "Minami Maki", Genres: []string{"Comedy", "Romance", "Shoujo"}, Status: "completed", TotalChapters: 99, Description: "Hikari strives to beat her rival Kei in academics and sports."},
		{ID: "snow-white-with-red-hair", Title: "Snow White with the Red Hair", Author: "Akiduki Sorata", Genres: []string{"Fantasy", "Romance", "Shoujo"}, Status: "ongoing", TotalChapters: 130, Description: "Shirayuki leaves her home and meets a prince while pursuing herbalism."},
		{ID: "high-school-debut", Title: "High School Debut", Author: "Kawahara Kazune", Genres: []string{"Romance", "Comedy", "Shoujo"}, Status: "completed", TotalChapters: 52, Description: "A former athlete seeks dating advice from a cool upperclassman."},
		{ID: "daytime-shooting-star", Title: "Daytime Shooting Star", Author: "Yamamori Mika", Genres: []string{"Romance", "Slice of Life", "Shoujo"}, Status: "completed", TotalChapters: 78, Description: "A country girl moves to Tokyo and falls in love for the first time."},
		{ID: "blue-spring-ride", Title: "Blue Spring Ride", Author: "Sakisaka Io", Genres: []string{"Romance", "Drama", "Shoujo"}, Status: "completed", TotalChapters: 49, Description: "Futaba attempts to reset her life in high school and meets her old crush."},

		// --- JOSEI (25) ---
		{ID: "chihayafuru", Title: "Chihayafuru", Author: "Suetsugu Yuki", Genres: []string{"Sports", "Drama", "Josei"}, Status: "completed", TotalChapters: 247, Description: "A girl strives to become the best Karuta player in the world."},
		{ID: "honey-and-clover", Title: "Honey and Clover", Author: "Umino Chica", Genres: []string{"Drama", "Romance", "Josei"}, Status: "completed", TotalChapters: 64, Description: "The lives and loves of five art college students living in the same building."},
		{ID: "nodame-cantabile", Title: "Nodame Cantabile", Author: "Ninomiya Tomoko", Genres: []string{"Music", "Comedy", "Josei"}, Status: "completed", TotalChapters: 136, Description: "A perfectionist conductor and a messy pianist help each other grow."},
		{ID: "kuragehime", Title: "Princess Jellyfish", Author: "Higashimura Akiko", Genres: []string{"Comedy", "Slice of Life", "Josei"}, Status: "completed", TotalChapters: 84, Description: "A jellyfish-obsessed girl lives in an apartment for geeky women."},
		{ID: "usagi-drop", Title: "Bunny Drop", Author: "Unita Yumi", Genres: []string{"Slice of Life", "Drama", "Josei"}, Status: "completed", TotalChapters: 62, Description: "A bachelor decides to raise his grandfather's illegitimate daughter."},
		{ID: "shouwa-genroku-rakugo", Title: "Shouwa Genroku Rakugo Shinjuu", Author: "Kumota Haruko", Genres: []string{"Drama", "Historical", "Josei"}, Status: "completed", TotalChapters: 28, Description: "The complex history of Rakugo performers across different generations."},
		{ID: "blue", Title: "Blue", Author: "Nanamonan Kiriko", Genres: []string{"Drama", "Romance", "Josei"}, Status: "completed", TotalChapters: 7, Description: "A sensitive exploration of the friendship and love between two schoolgirls."},
		{ID: "paradise-kiss", Title: "Paradise Kiss", Author: "Yazawa Ai", Genres: []string{"Fashion", "Romance", "Josei"}, Status: "completed", TotalChapters: 48, Description: "A student becomes a model for a group of fashion design students."},
		{ID: "blank-canvas", Title: "Blank Canvas", Author: "Higashimura Akiko", Genres: []string{"Autobiographical", "Arts", "Josei"}, Status: "completed", TotalChapters: 28, Description: "The author's real-life struggle to become a mangaka under a strict teacher."},
		{ID: "petshop-of-horrors", Title: "Pet Shop of Horrors", Author: "Akino Matsuri", Genres: []string{"Horror", "Supernatural", "Josei"}, Status: "completed", TotalChapters: 41, Description: "Count D sells rare pets with specific contracts that must not be broken."},
		{ID: "tokyo-tarareba-girls", Title: "Tokyo Tarareba Girls", Author: "Higashimura Akiko", Genres: []string{"Comedy", "Drama", "Josei"}, Status: "completed", TotalChapters: 28, Description: "Three women in their 30s wonder 'what if' they had made different choices."},
		{ID: "helter-skelter", Title: "Helter Skelter", Author: "Okazaki Kyoko", Genres: []string{"Psychological", "Drama", "Josei"}, Status: "completed", TotalChapters: 9, Description: "The physical and mental breakdown of a supermodel obsessed with beauty."},
		{ID: "07-ghost", Title: "07-Ghost", Author: "Amemiya Yuki", Genres: []string{"Fantasy", "Action", "Josei"}, Status: "completed", TotalChapters: 100, Description: "A former slave joins a church to escape a military empire and find his past."},
		{ID: "tramps-like-us", Title: "Tramps Like Us", Author: "Ogawa Yayoi", Genres: []string{"Romance", "Comedy", "Josei"}, Status: "completed", TotalChapters: 82, Description: "A successful career woman takes in a young man as a 'pet'."},
		{ID: "kids-on-the-slope", Title: "Kids on the Slope", Author: "Kodama Yuki", Genres: []string{"Music", "Drama", "Josei"}, Status: "completed", TotalChapters: 45, Description: "Two high schoolers in the 1960s bond over their love for jazz."},
		{ID: "don-t-say-mystery", Title: "Don't Say Mystery", Author: "Tamura Yumi", Genres: []string{"Mystery", "Psychological", "Josei"}, Status: "ongoing", TotalChapters: 50, Description: "A college student solves crimes through long philosophical monologues."},
		{ID: "hotaru-no-hikari", Title: "Hotaru no Hikari", Author: "Hiura Satoru", Genres: []string{"Comedy", "Romance", "Josei"}, Status: "completed", TotalChapters: 84, Description: "A woman who prefers to lounge at home accidentally lives with her boss."},
		{ID: "sakamichi-no-apollon", Title: "Sakamichi no Apollon", Author: "Kodama Yuki", Genres: []string{"Drama", "Music", "Josei"}, Status: "completed", TotalChapters: 45, Description: "A story of friendship and jazz in the late 1960s."},
		{ID: "hokusai-to-meshi-sae", Title: "Hokusai to Meshi Sae Are", Author: "Suzuki Sanami", Genres: []string{"Cooking", "Slice of Life", "Josei"}, Status: "completed", TotalChapters: 45, Description: "A poor student and her plushie find joy in cooking cheap, delicious meals."},
		{ID: "midnight-secretary", Title: "Midnight Secretary", Author: "Ohmi Tomu", Genres: []string{"Supernatural", "Romance", "Josei"}, Status: "completed", TotalChapters: 34, Description: "A secretary discovers her perfectionist boss is actually a vampire."},
		{ID: "bread-and-butter", Title: "Bread & Butter", Author: "Ashihara Hinako", Genres: []string{"Romance", "Slice of Life", "Josei"}, Status: "completed", TotalChapters: 52, Description: "An older woman and a baker find comfort and romance in a stationery shop."},
		{ID: "piece", Title: "Piece", Author: "Ashihara Hinako", Genres: []string{"Mystery", "Drama", "Josei"}, Status: "completed", TotalChapters: 34, Description: "A girl investigates the secret past of a classmate who recently died."},
		{ID: "suppli", Title: "Suppli", Author: "Okazaki Mari", Genres: []string{"Drama", "Romance", "Josei"}, Status: "completed", TotalChapters: 60, Description: "A woman struggles to balance her advertising career with her love life."},
		{ID: "karneval", Title: "Karneval", Author: "Mikanagi Touya", Genres: []string{"Fantasy", "Action", "Josei"}, Status: "completed", TotalChapters: 160, Description: "A boy searching for a friend joins a circus-themed defense organization."},
		{ID: "butterflies-flowers", Title: "Butterflies, Flowers", Author: "Yoshihara Yuki", Genres: []string{"Comedy", "Romance", "Josei"}, Status: "completed", TotalChapters: 40, Description: "A former rich girl works as a maid for the man who bought her house."},
	}

	stmt, _ := db.Prepare("INSERT OR IGNORE INTO manga (id, title, author, genres, status, total_chapters, description, cover_url) VALUES (?, ?, ?, ?, ?, ?, ?, ?)")
	defer stmt.Close()

	log.Println("Starting Manual Seed...")
	for _, entry := range entries {
		if existingTitles[strings.ToLower(entry.Title)] {
			continue // Skip if duplicate
		}

		log.Printf("Fetching cover for manual entry: %s", entry.Title)
		coverURL := fetchCoverForManualSeed(entry.Title)
		genresJSON, _ := json.Marshal(entry.Genres)

		_, err := stmt.Exec(entry.ID, entry.Title, entry.Author, string(genresJSON), entry.Status, entry.TotalChapters, entry.Description, coverURL)
		if err == nil {
			existingTitles[strings.ToLower(entry.Title)] = true // Mark as seen
		}
	}
}

// Function to fetch 25 entries per specified genre via the MangaDex API
func seedAPI(db *sql.DB) {
	demographics := []string{"shounen", "shoujo", "seinen", "josei"}

	stmt, _ := db.Prepare("INSERT OR IGNORE INTO manga (id, title, author, genres, status, total_chapters, description, cover_url) VALUES (?, ?, ?, ?, ?, ?, ?, ?)")
	defer stmt.Close()

	log.Println("Starting Dynamic API Seed...")
	for _, demo := range demographics {
		log.Printf("Fetching up to 25 unique %s series...", demo)

		// Use publicationDemographic filter and expand author & cover_art
		apiURL := fmt.Sprintf("https://api.mangadex.org/manga?publicationDemographic[]=%s&includes[]=author&includes[]=cover_art&limit=25", demo)
		resp, err := http.Get(apiURL)
		if err != nil || resp.StatusCode != 200 {
			log.Printf("Failed to fetch %s: %v", demo, err)
			continue
		}

		var dexResp MangaDexResponse
		if err := json.NewDecoder(resp.Body).Decode(&dexResp); err != nil {
			log.Printf("Failed to decode JSON for %s: %v", demo, err)
			resp.Body.Close()
			continue
		}
		resp.Body.Close()

		insertedCount := 0
		for _, item := range dexResp.Data {
			// Extract Title
			title := item.Attributes.Title["en"]
			if title == "" {
				for _, locTitle := range item.Attributes.Title {
					title = locTitle
					break
				}
			}
			if title == "" {
				title = "Unknown Title"
			}

			// Check Duplicate Map!
			if existingTitles[strings.ToLower(title)] {
				continue
			}

			// Extract Description
			desc := item.Attributes.Description["en"]
			if desc == "" {
				desc = "No description available."
			}

			// Extract Author and Cover Filename
			author := "Unknown Author"
			coverFileName := ""
			for _, rel := range item.Relationships {
				if rel.Type == "author" && rel.Attributes.Name != "" {
					author = rel.Attributes.Name
				} else if rel.Type == "cover_art" && rel.Attributes.FileName != "" {
					coverFileName = rel.Attributes.FileName
				}
			}

			// Format final Cover URL
			coverURL := "https://example.com/covers/default.jpg"
			if coverFileName != "" {
				coverURL = fmt.Sprintf("https://uploads.mangadex.org/covers/%s/%s", item.ID, coverFileName)
			}

			// Format genres
			genresJSON, _ := json.Marshal([]string{strings.Title(demo)})

			// Insert into DB
			_, err := stmt.Exec(item.ID, title, author, string(genresJSON), item.Attributes.Status, 0, desc, coverURL)
			if err == nil {
				existingTitles[strings.ToLower(title)] = true
				insertedCount++
			}
		}
		log.Printf("Successfully inserted %d unique %s series.", insertedCount, demo)
	}
}

func main() {
	log.Println("Initializing Database Connection...")
	database.InitDB("data.db")

	// 1. Run Manual Seeds
	seedManual(database.DB)

	// 2. Run Dynamic API Seeds
	seedAPI(database.DB)

	log.Println("Database Seeding Phase Complete!")
}
