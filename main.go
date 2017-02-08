package main;

import (
	"os"
	"fmt"
	"github.com/bwmarrin/discordgo"
	"encoding/binary"
	"io"
	"strings"
	"time"
	"math/rand"
	"io/ioutil"
	"path/filepath"
	"encoding/json"
	"runtime/debug"
	"regexp"
)

const DIRNAME = "Dank";
var sounds = make(map[string][][]byte, 0);
var images map[string]string;

var statuses = []string{
	"hidden object games",
	"Oh... Sir!",
	"Minecraft 1.0 ALPHA",
	"with your mother",
	"something",
	"something else",
	"bored",
	"dead"}

type Settings struct{
	playing bool
	commander string
}
var settings = make(map[string]*Settings);

func main(){
	args := os.Args[1:];

	if(len(args) < 1){
		fmt.Println("No token provided!");
		return;
	}
	token := args[0];

	fmt.Println("Loading...");

	err := os.MkdirAll(DIRNAME, 0755);
	if(err != nil){
		printErr(err);
		return;
	}
	files, err := ioutil.ReadDir(DIRNAME);
	if(err != nil){
		printErr(err);
		return;
	}
	for _, file := range files{
		if(file.IsDir()){
			continue;
		}
		name := file.Name();
		if(!strings.HasSuffix(name, ".dca")){
			continue;
		}

		bytes := make([][]byte, 0);
		err = load(name, &bytes);
		if(err != nil){
			continue;
		}
		
		name = strings.ToLower(strings.TrimSuffix(name, ".dca"));
		sounds[name] = bytes;
	}

	data, err := ioutil.ReadFile("Dank/images.json");
	if(err != nil){
		printErr(err);
		images = make(map[string]string, 0);
	} else {
		err = json.Unmarshal(data, &images);
		if(err != nil){
			printErr(err);
			images = make(map[string]string, 0);
		}
	}

	fmt.Println("Starting...");
	session, err := discordgo.New("Bot " + token);
	if(err != nil){
		printErr(err);
		return;
	}
	session.AddHandler(messageCreate);
	session.AddHandler(messageUpdate);
	err = session.Open();

	if(err != nil){
		printErr(err);
		return;
	}

	go func(){
		c := time.Tick(time.Second * 5);

		for _ = range c{
			err := session.UpdateStatus(0, statuses[rand.Intn(len(statuses))]);
			if(err != nil){ printErr(err); return; }
		}
	}();
	fmt.Println("Started!");

	<-make(chan struct{});
}

func load(file string, buffer *[][]byte) error{
	f, err := os.Open(filepath.Join(DIRNAME, file));
	defer f.Close();
	if err != nil {
		printErr(err);
		return err;
	}

	var length int16;
	for {
		err := binary.Read(f, binary.LittleEndian, &length);

		if(err == io.EOF || err == io.ErrUnexpectedEOF){
			break;
		} else if(err != nil){
			printErr(err);
			return err;
		}

		buf := make([]byte, length);
		err = binary.Read(f, binary.LittleEndian, &buf);
		if(err != nil){
			printErr(err);
			return err;
		}

		*buffer = append(*buffer, buf);
	}
	return nil;
}

func play(buffer [][]byte, session *discordgo.Session, guild, channel string, s *Settings){
	s.playing = true;
	vc, err := session.ChannelVoiceJoin(guild, channel, false, true);
	if(err != nil){
		printErr(err);
		s.playing = false;
		return;
	}

	err = vc.Speaking(true);
	if(err != nil){
		printErr(err);
		s.playing = false;
		return;
	}

	for _, buf := range buffer {
		if(!s.playing){ break; }
		vc.OpusSend <- buf;
	}

	err = vc.Speaking(false);
	if(err != nil){
		printErr(err);
	}
	err = vc.Disconnect();
	if(err != nil){
		printErr(err);
	}
	s.playing = false;
}

func messageCreate(session *discordgo.Session, event *discordgo.MessageCreate){
	message(session, event.Message)
}
func messageUpdate(session *discordgo.Session, event *discordgo.MessageUpdate){
	message(session, event.Message)
}
func message(session *discordgo.Session, event *discordgo.Message){
	if(event.Author == nil){ return; }
	msg := strings.ToLower(strings.TrimSpace(event.Content));
	author := event.Author;
	
	if(msg == ""){
		return;
	}

	channel, err := session.Channel(event.ChannelID);
	if(err != nil){ printErr(err); return; }

	if(channel.IsPrivate){
		return;
	}

	guild, err := session.Guild(channel.GuildID);
	if(err != nil){ printErr(err); return; }

	s := settings[guild.ID];
	if(s == nil){
		s = &Settings{};
		settings[guild.ID] = s;
	}

	if(s.commander != "" && s.commander != author.ID){
		return;
	}

	buffer, ok := sounds[msg];
	if(ok){
		if(!s.playing){
			for _, state := range guild.VoiceStates{
				if state.UserID == event.Author.ID{
					go react(session, event);
					play(buffer, session, guild.ID, state.ChannelID, s);
					return;
				}
			}
		}
		return;
	}

	for keyword, url := range images{
		contains, err := regexp.MatchString("(?i)\\b" +
			regexp.QuoteMeta(keyword) + "\\b", msg);
		if(err != nil){
			printErr(err);
			return;
		}
		if(contains){
			go react(session, event);
			_, err = session.ChannelMessageSendEmbed(event.ChannelID,
				&discordgo.MessageEmbed{
					Image: &discordgo.MessageEmbedImage{
						URL: url,
					},
				});
			if(err != nil){
				printErr(err);
			}
			return;
		}
	}

	switch(msg){
		case "thx":
			s.playing = false;
		case "listen only to me plz":
			s.commander = author.ID;
		case "every1 owns u stopad robot":
			s.commander = "";
	}
}

func react(session *discordgo.Session, event *discordgo.Message){
	err := session.MessageReactionAdd(event.ChannelID, event.ID, "👌");
	if(err != nil){
		printErr(err);
		return;
	}
	err = session.MessageReactionAdd(event.ChannelID, event.ID, "😂");
	if(err != nil){
		printErr(err);
	}
}

func printErr(err error){
	fmt.Fprintln(os.Stderr, "Error:", err);
	debug.PrintStack();
}
