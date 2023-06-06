# Takeout

## music cli

```bash
go run ./src/main.go music --help

Reorg the music files into a sane Artist/Album/Track structure

Usage:
  takeout music [flags]
  takeout music [command]

Available Commands:
  id3         read an mp3 file and output its id3 content
  id3         read an mp3 file and output its id3 content

Flags:
  -a, --analyze          print tracks analysis
  -h, --help             help for music
  -s, --save string      absolute path of where to save tracks
  -t, --trackMap         print trackMap detailed
  -b, --trackMapSimple   print trackMap simple print

Use "takeout music [command] --help" for more information about a command.
```

Example:

- Unzip / Untar and relocate all music files to `Takeout/YouTube and YouTube Music/music-uploads`

- run `go run ./src/main.go music -ab --save "$HOME/Music/YoutubeMusicOrg"  "$HOME/Music/Takeout/YouTube and YouTube Music/music-uploads"`

- `$HOME/Music/YoutubeMusicOrg` should now have all the found music sorted, anything missing should be relocated to `$HOME/Music/YoutubeMusicOrg/Unknown`
