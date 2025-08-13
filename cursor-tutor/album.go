type Album struct {
    ID     int  `json:"id"`
    Title  string  `json:"title"`
    Artist string  `json:"artist"`
    Price  float64 `json:"price"`
}

var albums = map[int]Album{
    1: {Title: "Thriller", Artist: "Michael Jackson", Price: 9.99},
    2: {Title: "Back in Black", Artist: "AC/DC", Price: 8.99},
    3: {Title: "The Dark Side of the Moon", Artist: "Pink Floyd", Price: 10.99},
    4: {Title: "Led Zeppelin IV", Artist: "Led Zeppelin", Price: 7.99},
    5: {Title: "Nevermind", Artist: "Nirvana", Price: 6.99},
    6: {Title: "Abbey Road", Artist: "The Beatles", Price: 11.99},
    7: {Title: "The Joshua Tree", Artist: "U2", Price: 9.99},
    8: {Title: "Appetite for Destruction", Artist: "Guns N' Roses", Price: 8.99},
    9: {Title: "The Wall", Artist: "Pink Floyd", Price: 12.99},
    10: {Title: "Ten", Artist: "Pearl Jam", Price: 7.99},
}

func findAlbumByID(id int, albums map[int]Album) *Album {
    album, ok := albums[id]
    if !ok {
        return nil
    }
    return &album
}
