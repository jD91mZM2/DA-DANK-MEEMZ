package main

import (
	"os"
	"fmt"
	"github.com/bwmarrin/discordgo"
	"encoding/binary"
	"io"
	"strings"
	"time"
	"math/rand"
)

var JohnCena = make([][]byte, 0);
var Elevator = make([][]byte, 0);
var Rickroll = make([][]byte, 0);
var Letter = make([][]byte, 0);
var Cri = make([][]byte, 0);
var NumberHat = make([][]byte, 0);
var ExoticButters = make([][]byte, 0);
var DamnSon = make([][]byte, 0);
var Jeff = make([][]byte, 0);
var Nigga = make([][]byte, 0);
var RussianSinger = make([][]byte, 0);
var SadViolin = make([][]byte, 0);
var ShutUp = make([][]byte, 0);
var Triple = make([][]byte, 0);
var TurTur = make([][]byte, 0);
var Weed = make([][]byte, 0);
var XFiles = make([][]byte, 0);

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
	vc *discordgo.VoiceConnection
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

	load("John Cena", &JohnCena);
	load("Elevator", &Elevator);
	load("Rickroll", &Rickroll);
	load("Cri", &Cri);
	load("Letter", &Letter);
	load("NumberHat", &NumberHat);
	load("ExoticButters", &ExoticButters);
	load("damnson", &DamnSon);
	load("jeff", &Jeff);
	load("nigga", &Nigga);
	load("russianSinger", &RussianSinger);
	load("sadviolin", &SadViolin);
	load("shutup", &ShutUp);
	load("triple", &Triple);
	load("turtur", &TurTur);
	load("weed", &Weed);
	load("xfiles", &XFiles);

	fmt.Println("Starting...");
	d, err := discordgo.New("Bot " + token);
	if(err != nil){
		fmt.Fprintln(os.Stderr, "I THINK THERE WAS ERROR ", err);
		return;
	}
	d.AddHandler(ready);
	d.AddHandler(messageCreate);
	err = d.Open();

	if(err != nil){
		fmt.Fprintln(os.Stderr, "I THINK THERE WAS ERROR ", err);
		return;
	}
	fmt.Println("Started!");

	<-make(chan struct{});
}

func load(file string, buffer *[][]byte){
	f, err := os.Open("Dank/" + file + ".dca");
	if err != nil {
		fmt.Fprintln(os.Stderr, "FILE WAS WEIRD IDK: ", err);
		return;
	}

	var length int16;
	for {
		err := binary.Read(f, binary.LittleEndian, &length);

		if(err == io.EOF || err == io.ErrUnexpectedEOF){
			return;
		} else if(err != nil) {
			fmt.Fprintln(os.Stderr, "IDK WAT U DID BUT WEL DONN NOOB, ", err);
			return;
		}

		buf := make([]byte, length);
		binary.Read(f, binary.LittleEndian, &buf);

		*buffer = append(*buffer, buf);
	}
}

func play(buffer [][]byte, session *discordgo.Session, guild, channel string, s *Settings){
	s.playing = true;
	var err error;
	s.vc, err = session.ChannelVoiceJoin(guild, channel, false, true);
	if(err != nil){
		fmt.Fprintln(os.Stderr, "LEL FREGGIN NOOB, ", err);
		s.playing = false;
		return;
	}

	err = s.vc.Speaking(true);
	if(err != nil){
		fmt.Fprintln(os.Stderr, "I HAS TO TYPE ALL DESE ERROR MESSAGES " +
		"HAVE SOME SYMPATHY PLS ", err);
		return;
	}

	for _, buf := range buffer {
		if s.vc == nil { return; }
		s.vc.OpusSend <- buf;
	}

	err = s.vc.Speaking(false);
	if(err != nil){
		fmt.Fprintln(os.Stderr, err);
		return;
	}
	err = s.vc.Disconnect();
	if(err != nil){
		fmt.Fprintln(os.Stderr, err);
		return;
	}
	s.playing = false;
}

func messageCreate(session *discordgo.Session, event *discordgo.MessageCreate){
	msg := strings.ToLower(strings.TrimSpace(event.Content));
	author := event.Author;

	channel, err := session.State.Channel(event.ChannelID);
	if(err != nil){ fmt.Fprintln(os.Stderr, err); return; }

	guild, err := session.State.Guild(channel.GuildID);
	if(err != nil){ fmt.Fprintln(os.Stderr, err); return; }
	//member, err := session.State.Member(guild.ID, author.ID);

	s := settings[guild.ID];
	if(s == nil){
		s = &Settings{};
		settings[guild.ID] = s;
	}

	if(s.commander != "" && s.commander != author.ID){
		return;
	}

	var buffer [][]byte = nil;
	switch(msg){
		case "john cena":		buffer = JohnCena;
		case "waiting":			buffer = Elevator;
		case "rickroll":		buffer = Rickroll;
		case "cri":				buffer = Cri;
		case "letter":			buffer = Letter;
		case "numbr hat":		buffer = NumberHat;
		case "exotic butters":	buffer = ExoticButters;
		case "damn son":		buffer = DamnSon;
		case "jeff":			buffer = Jeff;
		case "nigga":			buffer = Nigga;
		case "russian singer":	buffer = RussianSinger;
		case "sad violin":		buffer = SadViolin;
		case "shut up":			buffer = ShutUp;
		case "triple":			buffer = Triple;
		case "turtur":			buffer = TurTur;
		case "weed":			buffer = Weed;
		case "illuminati":		buffer = XFiles;
		case "thx":
			if(s.vc != nil){
				err := s.vc.Speaking(false);
				if(err != nil){ fmt.Fprintln(os.Stderr, err); return; }

				err = s.vc.Disconnect();
				if(err != nil){ fmt.Fprintln(os.Stderr, err); return; }

				s.playing = false;
			}
		case "listen only to me plz":
			s.commander = author.ID;
		case "every1 owns u stopad robot":
			s.commander = "";
	}

	if(buffer != nil && !s.playing){
		for _, state := range guild.VoiceStates{
			if state.UserID == event.Author.ID{
				play(buffer, session, guild.ID, state.ChannelID, s);
			}
		}
	}
}

func ready(session *discordgo.Session, event *discordgo.Ready){
	c := time.Tick(time.Second * 5);

	for _ = range c{
		err := session.UpdateStatus(0, statuses[rand.Intn(len(statuses))]);
		if(err != nil){ fmt.Fprintln(os.Stderr, err); return; }
	}
}
