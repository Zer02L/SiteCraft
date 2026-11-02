package cinematic_2

// NavLink represents a navigation item in the header
type NavLink struct {
	Label    string
	URL      string
	External bool
}

// HeroProps holds configuration for Screen 1 (Cinematic 3D Character Artist Hero)
type HeroProps struct {
	BrandName    string
	Subtitle     string
	VideoURL     string
	PosterURL    string
	NavLinks     []NavLink
	CartCount    int
	TargetScroll string
	AriaLabel    string
	Class        string // Optional container CSS overrides (e.g. "bg-surface text-content")
}

// CollageItem represents a single portfolio artwork card in Screen 2
type CollageItem struct {
	Title    string
	ImageURL string
	Alt      string
	Link     string
}

// CollageProps holds configuration for Screen 2 (Rotated Infinite Marquee Collage)
type CollageProps struct {
	ID          string
	Title       string
	TitleLink   string
	RotationDeg float64
	Row1        []CollageItem
	Row2        []CollageItem
	Row3        []CollageItem
	Class       string // Optional container CSS overrides (e.g. "bg-surface text-content")
	CardRadius  string // Optional card radius token (defaults to "rounded-xl sm:rounded-2xl")
}

// ShowcaseProps bundles both screens together
type ShowcaseProps struct {
	Hero    HeroProps
	Collage CollageProps
}

// DefaultHeroProps returns canonical data for JHill Screen 1
func DefaultHeroProps() HeroProps {
	return HeroProps{
		BrandName:    "JHILL",
		Subtitle:     "3D Character Artist",
		VideoURL:     "https://www.jhill.com/cdn/shop/videos/c/vp/01ed761ace7848c4a9bdf42f60c9b04b/01ed761ace7848c4a9bdf42f60c9b04b.HD-1080p-7.2Mbps-59074109.mp4?v=0",
		PosterURL:    "https://www.jhill.com/cdn/shop/files/preview_images/01ed761ace7848c4a9bdf42f60c9b04b.thumbnail.0000000000.jpg?v=1759899190&width=1920",
		TargetScroll: "#portfolio",
		CartCount:    0,
		NavLinks: []NavLink{
			{Label: "Portfolio", URL: "#portfolio", External: false},
			{Label: "Resources", URL: "https://www.jhill.com/collections/all", External: true},
			{Label: "YouTube", URL: "https://www.youtube.com/@JHill3D", External: true},
			{Label: "Discord", URL: "https://discord.gg/jhill", External: true},
			{Label: "About", URL: "https://www.artstation.com/jhill/profile", External: true},
		},
	}
}

// DefaultCollageProps returns canonical portfolio artwork rows for JHill Screen 2
func DefaultCollageProps() CollageProps {
	return CollageProps{
		ID:          "portfolio",
		Title:       "Portfolio",
		TitleLink:   "https://www.artstation.com/jhill",
		RotationDeg: 12.0,
		Row1: []CollageItem{
			{
				Title:    "Mom - Back 4 Blood",
				ImageURL: "https://www.jhill.com/cdn/shop/files/j-hill-mom-game-portrait-01.jpg?v=1758256340",
				Alt:      "3D character portrait of Mom from Back 4 Blood by JHill",
			},
			{
				Title:    "Focus Portrait Split",
				ImageURL: "https://www.jhill.com/cdn/shop/files/j-hill-focus-portrait-split.jpg?v=1758256338",
				Alt:      "Facial topology and skin shading split render by JHill",
			},
			{
				Title:    "Hoffman - Back 4 Blood",
				ImageURL: "https://www.jhill.com/cdn/shop/files/j-hill-hoffman-game-portrait-2.jpg?v=1758256340",
				Alt:      "3D character model of Hoffman by JHill",
			},
			{
				Title:    "Hunter Beauty",
				ImageURL: "https://www.jhill.com/cdn/shop/files/j-hill-hunter-beauty-square-3_0d2e52d6-cf7e-463a-8811-f71fd2638b72.jpg?v=1758711337",
				Alt:      "High poly hunter character beauty render",
			},
			{
				Title:    "Cyberpunk V",
				ImageURL: "https://www.jhill.com/cdn/shop/files/j-hill-v-01.jpg?v=1758256339",
				Alt:      "V character model render from Cyberpunk",
			},
			{
				Title:    "Steve Portrait",
				ImageURL: "https://www.jhill.com/cdn/shop/files/j-hill-steve-portrait-01.jpg?v=1758256338",
				Alt:      "Realistic human character head study by JHill",
			},
			{
				Title:    "Evangelo",
				ImageURL: "https://www.jhill.com/cdn/shop/files/j-hill-evangelo-game-portrait-03_2fe56f09-5ef7-41be-8fb5-516388f5dd7a.jpg?v=1758711337",
				Alt:      "Evangelo game ready character model",
			},
		},
		Row2: []CollageItem{
			{
				Title:    "Nima Close",
				ImageURL: "https://www.jhill.com/cdn/shop/files/j-hill-nima-close.jpg?v=1758256339",
				Alt:      "Nima cinematic portrait close up",
			},
			{
				Title:    "Orc Creature",
				ImageURL: "https://www.jhill.com/cdn/shop/files/j-hill-orc-comp-5-3.jpg?v=1758256340",
				Alt:      "3D sculpted fantasy Orc creature",
			},
			{
				Title:    "Bloodhound - Apex Legends",
				ImageURL: "https://www.jhill.com/cdn/shop/files/j-hill-bloodhound-portrait_ebe021e0-2f5e-45ae-9a6d-a35c3a2ab528.jpg?v=1758711337",
				Alt:      "Apex Legends Bloodhound portrait",
			},
			{
				Title:    "Cybernetic Sculpt",
				ImageURL: "https://www.jhill.com/cdn/shop/files/j-hill-2_b5df73a9-0433-4a5d-98cd-1c1ffeefcf6c.jpg?v=1758711337",
				Alt:      "Cybernetic biomechanical character sculpt",
			},
			{
				Title:    "Walker - Back 4 Blood",
				ImageURL: "https://www.jhill.com/cdn/shop/files/j-hill-walker-game-portrait-1.jpg?v=1758256340",
				Alt:      "Walker game portrait",
			},
			{
				Title:    "Marmoset Toolbag Beauty",
				ImageURL: "https://www.jhill.com/cdn/shop/files/jay-hill-marmoset-beauty.jpg?v=1758256339",
				Alt:      "Realtime lighting in Marmoset Toolbag",
			},
			{
				Title:    "Thanos Web Study",
				ImageURL: "https://www.jhill.com/cdn/shop/files/j-hill-thanos-portrait-web.jpg?v=1758256339",
				Alt:      "Thanos realistic portrait sculpt",
			},
		},
		Row3: []CollageItem{
			{
				Title:    "Cyberpunk Color Full",
				ImageURL: "https://www.jhill.com/cdn/shop/files/j-hill-cyberpunk-color-full_914a7096-a729-4c19-9137-6efa6a78d0b2.jpg?v=1758711337",
				Alt:      "Full body Cyberpunk character render",
			},
			{
				Title:    "Doc - Back 4 Blood",
				ImageURL: "https://www.jhill.com/cdn/shop/files/j-hill-doc-game-portrait-full.jpg?v=1758256340",
				Alt:      "Doc full character portrait",
			},
			{
				Title:    "Judge Dredd",
				ImageURL: "https://www.jhill.com/cdn/shop/files/jay-hill-dredd-render.jpg?v=1758256339",
				Alt:      "Judge Dredd character 3D sculpt",
			},
			{
				Title:    "Holly - Back 4 Blood",
				ImageURL: "https://www.jhill.com/cdn/shop/files/j-hill-holly-game-portrait-1.jpg?v=1758256339",
				Alt:      "Holly game model portrait",
			},
			{
				Title:    "Werewolf Creature",
				ImageURL: "https://www.jhill.com/cdn/shop/files/j-hill-werewolf-web.jpg?v=1758256339",
				Alt:      "Werewolf creature anatomy sculpt",
			},
			{
				Title:    "Night Queen",
				ImageURL: "https://www.jhill.com/cdn/shop/files/j-hill-nightqueen-comp.jpg?v=1758256339",
				Alt:      "Night Queen composite render",
			},
			{
				Title:    "Sauron Beauty Front",
				ImageURL: "https://www.jhill.com/cdn/shop/files/j-hill-sauron-beauty-front-dof.jpg?v=1758256340",
				Alt:      "Sauron armor and helmet beauty shot",
			},
		},
	}
}
