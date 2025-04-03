package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"math/rand"
	"mime/multipart"
	"net/http"
	"net/http/cookiejar"
	"net/url"
	"strings"
	"time"

	"github.com/tiktoken-go/tokenizer"
)

var (
	defaultTextModel           = "anthropic-claude-3-7-sonnet-latest"
	defaultPricePerTokenInput  = float64(3e-06)
	defaultPricePerTokenOutput = float64(1.5e-05)

	minSleep = 30000 // 3 mins
	maxSleep = 90000 // 9 mins

	errMinSleep = 1800000 // 30 mins
	errMaxSleep = 5400000 // 90 mins
)

var topics = []string{
	"Abandoned theme park",
	"Alien invasion",
	"Alternative history",
	"Ancient civilization",
	"Artificial intelligence",
	"Assassination plot",
	"Astronaut stranded on Mars",
	"Bank heist",
	"Biological disaster",
	"Boarding school mystery",
	"Coming of age",
	"Conspiracy theory",
	"Cryptid sightings",
	"Cyberpunk city",
	"Detective solving a murder",
	"Dinosaur theme park",
	"Disaster movie",
	"Dystopian future",
	"Environmental disaster",
	"Epic fantasy quest",
	"Escape from prison",
	"Fantasy world war",
	"First contact with aliens",
	"Futuristic sports",
	"Ghost story",
	"Haunted house",
	"Historical romance",
	"Intergalactic politics",
	"Island survival",
	"Jungle adventure",
	"Lost city",
	"Love triangle",
	"Magical academy",
	"Medical thriller",
	"Medieval fantasy",
	"Mystery cult",
	"Natural disaster",
	"Ninja warrior",
	"Parallel universe",
	"Philosophical debate",
	"Pizza delivery guy gets caught up in a crime",
	"Pirate story",
	"Post-apocalyptic wasteland",
	"Professional wrestling",
	"Psychological horror",
	"Punk rock rebellion",
	"Quest for a legendary artifact",
	"Reality TV show",
	"Road trip across the United States",
	"Robot uprising",
	"Romantic comedy",
	"Scientific experiment gone wrong",
	"Sea monster",
	"Sherlock Holmes-style mystery",
	"Space exploration",
	"Space station disaster",
	"Special forces operation",
	"Steampunk city",
	"Superhero origin story",
	"Survival on a deserted island",
	"Sword and sorcery",
	"Terrorist plot",
	"The last human on Earth",
	"Time travel",
	"Treasure hunt",
	"Underwater city",
	"Urban fantasy",
	"Utopian society",
	"Vampire hunter",
	"Virtual reality world",
	"War between humans and AI",
	"Werewolf pack",
	"Western frontier",
	"Witches' coven",
	"Zombie apocalypse",
	"Zoo escape",
	"Superhero team",
	"Historical fiction",
	"Mythological creatures",
	"Ancient mythology",
	"Small town secrets",
	"Creepy carnival",
	"High school drama",
	"Food critic",
	"Road trip in a foreign country",
	"Superhero romance",
	"Undercover cop",
	"Mystery at a summer camp",
	"Catastrophe on a spaceship",
	"Disaster in the wilderness",
	"Cthulhu mythos",
	"Ancient curses",
	"Abandoned asylum",
	"Small town legend",
	"Space battle",
	"Future world",
	"High-tech mystery",
	"Supernatural romance",
	"Teenager with a superpower",
	"Fighting against an oppressive government",
	"Group of friends on a mission",
	"Ancient alien technology",
	"Lost in a foreign country",
	"Mysterious library",
	"Gang war",
	"Cthulhu-inspired story",
	"Time loop",
	"Trapped in a never-ending dream",
	"Fighting against a corrupt corporation",
	"Haunted by the past",
	"Chased by a supernatural entity",
	"Trapped in a time loop",
	"Deserted research station",
	"Escape from a strange and fantastical world",
	"Chasing a legendary creature",
	"Lost in a supernatural realm",
	"Haunted by guilt",
	"Chosen one destined for greatness",
	"Trapped in a mysterious labyrinth",
	"Surviving in a deadly game",
	"Frozen in a winter wonderland",
	"Lost at sea with no direction",
	"Haunted by a cursed object",
	"Pursued by a demonic entity",
	"Trapped in a supernatural world",
	"Chased by a malevolent spirit",
	"Captured by an alien species",
	"Abandoned on a desolate planet",
	"Fighting for survival in a post-apocalyptic wasteland",
	"Frozen in time",
	"Lost in a mystical forest",
	"Haunted by the spirit of a deceased loved one",
	"Chasing a mysterious figure",
	"Trapped in a labyrinthine library",
	"Surviving a zombie apocalypse",
	"Chosen to wield an ancient power",
	"Fighting against an otherworldly foe",
	"Chased by a monstrous creature",
	"Frozen in fear of the unknown",
	"Trapped in a world of magic and mayhem",
	"Abandoned in an apocalyptic wasteland",
	"Haunted by the ghosts of the past",
	"Chasing a mysterious prophecy",
	"Lost in a labyrinth of mirrors",
	"Frozen in a world of ice and snow",
	"Trapped in a supernatural realm",
	"Chosen to fulfill an ancient destiny",
	"Pursued by a powerful and malevolent force",
	"Fighting for survival in a futuristic world",
	"Frozen in a moment of time",
	"Trapped in a cycle of reincarnation",
	"Haunted by the vengeful dead",
	"Chosen to wield the power of the elements",
	"Abandoned on a deserted island",
	"Lost in a maze of memories",
	"Chasing a legendary treasure",
	"Trapped in a world of nightmares",
	"Fighting against an evil empire",
	"Frozen in a world of perpetual darkness",
	"Haunted by the horrors of the past",
	"Pursued by a powerful and ruthless organization",
	"Trapped in a time loop of terror",
	"Chosen to wield the power of magic",
	"Abandoned in a world of wonder",
	"Lost in a sea of secrets and lies",
	"Fighting for survival in a world of fantasy",
	"Frozen in a moment of madness",
	"Trapped in a world of illusions",
	"Haunted by the ghosts of war",
	"Chasing a mysterious and elusive figure",
	"Pursued by a monstrous and malevolent creature",
	"Fighting against an army of the undead",
	"Trapped in a supernatural trap",
	"Abandoned in a desolate and barren land",
	"Lost in a labyrinth of forgotten lore",
	"Chosen to wield the power of the gods",
	"Frozen in a world of fear and uncertainty",
	"Haunted by the horrors of the unknown",
	"Pursued by a powerful and relentless enemy",
	"Trapped in a cycle of chaos and destruction",
	"Fighting for survival in a world of wonders",
	"Abandoned in a world of darkness and despair",
	"Lost in a sea of sorrow and suffering",
	"Chosen to wield the power of the elements",
	"Frozen in a moment of desperation",
	"Trapped in a world of mystery and intrigue",
	"Haunted by the ghosts of the past",
	"Pursued by a mysterious and malevolent force",
	"Fighting for survival in a post-apocalyptic world",
	"Trapped in a world of supernatural terror",
	"Abandoned in a desolate and barren wasteland",
	"Lost in a labyrinth of forgotten knowledge",
	"Chosen to wield the power of prophecy",
	"Frozen in a world of ice and darkness",
	"Haunted by the horrors of the apocalypse",
	"Pursued by a powerful and ruthless enemy",
	"Trapped in a time loop of terror and despair",
	"Fighting for survival in a world of fantasy and horror",
	"Abandoned in a world of wonder and magic",
	"Lost in a sea of secrets and deception",
	"Chosen to wield the power of the gods",
	"Frozen in a moment of madness and despair",
	"Trapped in a supernatural trap",
	"Haunted by the ghosts of war and violence",
	"Pursued by a monstrous and malevolent creature",
	"Fighting against an army of the undead and the forces of darkness",
	"Trapped in a cycle of chaos and destruction",
	"Abandoned in a desolate and barren land",
	"Lost in a labyrinth of forgotten lore and legend",
	"Chosen to wield the power of prophecy and fate",
	"Frozen in a world of fear and uncertainty",
	"Haunted by the horrors of the unknown and the supernatural",
	"Pursued by a powerful and relentless enemy",
	"Trapped in a world of mystery and intrigue",
	"Fighting for survival in a world of fantasy and science fiction",
	"Abandoned in a world of darkness and despair",
	"Lost in a sea of sorrow and suffering",
	"Chosen to wield the power of the elements and the forces of nature",
	"Frozen in a moment of desperation and hopelessness",
	"Trapped in a cycle of chaos and destruction",
	"Haunted by the ghosts of the past and the horrors of the present",
	"Pursued by a mysterious and malevolent force",
	"Fighting for survival in a post-apocalyptic world",
	"Trapped in a world of supernatural terror and horror",
	"Abandoned in a desolate and barren wasteland",
	"Lost in a labyrinth of forgotten knowledge and ancient lore",
	"Chosen to wield the power of prophecy and destiny",
	"Frozen in a world of ice and darkness",
	"Haunted by the horrors of the apocalypse and the forces of darkness",
	"Pursued by a powerful and ruthless enemy",
	"Trapped in a time loop of terror and despair",
	"Fighting for survival in a world of fantasy and horror",
	"Abandoned in a world of wonder and magic",
	"Lost in a sea of secrets and deception",
	"Chosen to wield the power of the gods and the forces of nature",
	"Frozen in a moment of madness and despair",
	"Trapped in a supernatural trap",
	"Haunted by the ghosts of war and violence",
	"Pursued by a monstrous and malevolent creature",
	"Fighting against an army of the undead and the forces of darkness",
	"Trapped in a cycle of chaos and destruction",
	"Abandoned in a desolate and barren land",
	"Lost in a labyrinth of forgotten lore and legend",
	"Chosen to wield the power of prophecy and fate",
	"Frozen in a world of fear and uncertainty",
	"Haunted by the horrors of the unknown and the supernatural",
	"Pursued by a powerful and relentless enemy",
	"Trapped in a world of mystery and intrigue",
	"Fighting for survival in a world of fantasy and science fiction",
	"Abandoned in a world of darkness and despair",
	"Lost in a sea of sorrow and suffering",
	"Chosen to wield the power of the elements and the forces of nature",
	"Frozen in a moment of desperation and hopelessness",
	"Trapped in a cycle of chaos and destruction",
	"Haunted by the ghosts of the past and the horrors of the present",
	"Pursued by a mysterious and malevolent force",
	"Fighting for survival in a post-apocalyptic world",
	"Trapped in a world of supernatural terror and horror",
	"Abandoned in a desolate and barren wasteland",
	"Lost in a labyrinth of forgotten knowledge and ancient lore",
	"Chosen to wield the power of prophecy and destiny",
	"Frozen in a world of ice and darkness",
	"Haunted by the horrors of the apocalypse and the forces of darkness",
	"Pursued by a powerful and ruthless enemy",
	"Trapped in a time loop of terror and despair",
	"Fighting for survival in a world of fantasy and horror",
	"Abandoned in a world of wonder and magic",
	"Lost in a sea of secrets and deception",
	"Chosen to wield the power of the gods and the forces of nature",
	"Frozen in a moment of madness and despair",
	"Trapped in a supernatural trap",
	"Haunted by the ghosts of war and violence",
	"Pursued by a monstrous and malevolent creature",
	"Fighting against an army of the undead and the forces of darkness",
	"Trapped in a cycle of chaos and destruction",
	"Abandoned in a desolate and barren land",
	"Lost in a labyrinth of forgotten lore and legend",
	"Chosen to wield the power of prophecy and fate",
	"Frozen in a world of fear and uncertainty",
	"Haunted by the horrors of the unknown and the supernatural",
	"Pursued by a powerful and relentless enemy",
	"Trapped in a world of mystery and intrigue",
	"Fighting for survival in a world of fantasy and science fiction",
	"Abandoned in a world of darkness and despair",
	"Lost in a sea of sorrow and suffering",
	"Chosen to wield the power of the elements and the forces of nature",
	"Frozen in a moment of desperation and hopelessness",
	"Trapped in a cycle of chaos and destruction",
	"Haunted by the ghosts of the past and the horrors of the present",
	"Pursued by a mysterious and malevolent force",
	"Fighting for survival in a post-apocalyptic world",
	"Trapped in a world of supernatural terror and horror",
	"Abandoned in a desolate and barren wasteland",
	"Lost in a labyrinth of forgotten knowledge and ancient lore",
	"Chosen to wield the power of prophecy and destiny",
	"Frozen in a world of ice and darkness",
	"Haunted by the horrors of the apocalypse and the forces of darkness",
	"Pursued by a powerful and ruthless enemy",
	"Trapped in a time loop of terror and despair",
	"Fighting for survival in a world of fantasy and horror",
	"Abandoned in a world of wonder and magic",
	"Lost in a sea of secrets and deception",
	"Chosen to wield the power of the gods and the forces of nature",
	"Frozen in a moment of madness and despair",
	"Trapped in a supernatural trap",
	"Haunted by the ghosts of war and violence",
	"Pursued by a monstrous and malevolent creature",
	"Fighting against an army of the undead and the forces of darkness",
	"Trapped in a cycle of chaos and destruction",
	"Abandoned in a desolate and barren land",
	"Lost in a labyrinth of forgotten lore and legend",
	"Chosen to wield the power of prophecy and fate",
	"Frozen in a world of fear and uncertainty",
	"Haunted by the horrors of the unknown and the supernatural",
	"Pursued by a powerful and relentless enemy",
	"Trapped in a world of mystery and intrigue",
	"Fighting for survival in a world of fantasy and science fiction",
	"Abandoned in a world of darkness and despair",
	"Lost in a sea of sorrow and suffering",
	"Chosen to wield the power of the elements and the forces of nature",
	"Frozen in a moment of desperation and hopelessness",
	"Trapped in a cycle of chaos and destruction",
	"Haunted by the ghosts of the past and the horrors of the present",
	"Pursued by a mysterious and malevolent force",
	"Fighting for survival in a post-apocalyptic world",
	"Trapped in a world of supernatural terror and horror",
	"Abandoned in a desolate and barren wasteland",
	"Lost in a labyrinth of forgotten knowledge and ancient lore",
	"Chosen to wield the power of prophecy and destiny",
	"Frozen in a world of ice and darkness",
	"Haunted by the horrors of the apocalypse and the forces of darkness",
	"Pursued by a powerful and ruthless enemy",
	"Trapped in a time loop of terror and despair",
	"Fighting for survival in a world of fantasy and horror",
	"Abandoned in a world of wonder and magic",
	"Lost in a sea of secrets and deception",
	"Chosen to wield the power of the gods and the forces of nature",
	"Frozen in a moment of madness and despair",
	"Trapped in a supernatural trap",
	"Haunted by the ghosts of war and violence",
	"Pursued by a monstrous and malevolent creature",
	"Fighting against an army of the undead and the forces of darkness",
	"Trapped in a cycle of chaos and destruction",
	"Abandoned in a desolate and barren land",
	"Lost in a labyrinth of forgotten lore and legend",
	"Chosen to wield the power of prophecy and fate",
	"Frozen in a world of fear and uncertainty",
	"Haunted by the horrors of the unknown and the supernatural",
	"Pursued by a powerful and relentless enemy",
	"Trapped in a world of mystery and intrigue",
	"Fighting for survival in a world of fantasy and science fiction",
	"Abandoned in a world of darkness and despair",
	"Lost in a sea of sorrow and suffering",
	"Chosen to wield the power of the elements and the forces of nature",
	"Frozen in a moment of desperation and hopelessness",
	"Trapped in a cycle of chaos and destruction",
	"Haunted by the ghosts of the past and the horrors of the present",
	"Pursued by a mysterious and malevolent force",
	"Fighting for survival in a post-apocalyptic world",
	"Trapped in a world of supernatural terror and horror",
	"Abandoned in a desolate and barren wasteland",
	"Lost in a labyrinth of forgotten knowledge and ancient lore",
	"Chosen to wield the power of prophecy and destiny",
	"Frozen in a world of ice and darkness",
	"Haunted by the horrors of the apocalypse and the forces of darkness",
	"Pursued by a powerful and ruthless enemy",
	"Trapped in a time loop of terror and despair",
	"Fighting for survival in a world of fantasy and horror",
	"Abandoned in a world of wonder and magic",
	"Lost in a sea of secrets and deception",
	"Chosen to wield the power of the gods and the forces of nature",
	"Frozen in a moment of madness and despair",
	"Trapped in a supernatural trap",
	"Haunted by the ghosts of war and violence",
	"Pursued by a monstrous and malevolent creature",
	"Fighting against an army of the undead and the forces of darkness",
	"Trapped in a cycle of chaos and destruction",
	"Abandoned in a desolate and barren land",
	"Lost in a labyrinth of forgotten lore and legend",
	"Chosen to wield the power of prophecy and fate",
	"Frozen in a world of fear and uncertainty",
	"Haunted by the horrors of the unknown and the supernatural",
	"Pursued by a powerful and relentless enemy",
	"Trapped in a world of mystery and intrigue",
	"Fighting for survival in a world of fantasy and science fiction",
	"Abandoned in a world of darkness and despair",
	"Lost in a sea of sorrow and suffering",
	"Chosen to wield the power of the elements and the forces of nature",
	"Frozen in a moment of desperation and hopelessness",
	"Trapped in a cycle of chaos and destruction",
	"Haunted by the ghosts of the past and the horrors of the present",
	"Pursued by a mysterious and malevolent force",
	"Fighting for survival in a post-apocalyptic world",
	"Trapped in a world of supernatural terror and horror",
	"Abandoned in a desolate and barren wasteland",
	"Lost in a labyrinth of forgotten knowledge and ancient lore",
	"Chosen to wield the power of prophecy and destiny",
	"Frozen in a world of ice and darkness",
	"Haunted by the horrors of the apocalypse and the forces of darkness",
	"Pursued by a powerful and ruthless enemy",
	"Trapped in a time loop of terror and despair",
	"Fighting for survival in a world of fantasy and horror",
	"Abandoned in a world of wonder and magic",
	"Lost in a sea of secrets and deception",
	"Chosen to wield the power of the gods and the forces of nature",
	"Frozen in a moment of madness and despair",
	"Trapped in a supernatural trap",
	"Haunted by the ghosts of war and violence",
	"Pursued by a monstrous and malevolent creature",
	"Fighting against an army of the undead and the forces of darkness",
	"Trapped in a cycle of chaos and destruction",
	"Abandoned in a desolate and barren land",
	"Lost in a labyrinth of forgotten lore and legend",
	"Chosen to wield the power of prophecy and fate",
	"Frozen in a world of fear and uncertainty",
	"Haunted by the horrors of the unknown and the supernatural",
	"Pursued by a powerful and relentless enemy",
	"Trapped in a world of mystery and intrigue",
	"Fighting for survival in a world of fantasy and science fiction",
	"Abandoned in a world of darkness and despair",
	"Lost in a sea of sorrow and suffering",
	"Chosen to wield the power of the elements and the forces of nature",
	"Frozen in a moment of desperation and hopelessness",
	"Trapped in a cycle of chaos and destruction",
	"Haunted by the ghosts of the past and the horrors of the present",
	"Pursued by a mysterious and malevolent force",
	"Fighting for survival in a post-apocalyptic world",
	"Trapped in a world of supernatural terror and horror",
	"Abandoned in a desolate and barren wasteland",
	"Lost in a labyrinth of forgotten knowledge and ancient lore",
	"Chosen to wield the power of prophecy and destiny",
	"Frozen in a world of ice and darkness",
	"Haunted by the horrors of the apocalypse and the forces of darkness",
	"Pursued by a powerful and ruthless enemy",
	"Trapped in a time loop of terror and despair",
	"Fighting for survival in a world of fantasy and horror",
	"Abandoned in a world of wonder and magic",
	"Lost in a sea of secrets and deception",
	"Chosen to wield the power of the gods and the forces of nature",
	"Frozen in a moment of madness and despair",
	"Trapped in a supernatural trap",
	"Haunted by the ghosts of war and violence",
	"Pursued by a monstrous and malevolent creature",
	"Fighting against an army of the undead and the forces of darkness",
	"Trapped in a cycle of chaos and destruction",
	"Abandoned in a desolate and barren land",
	"Lost in a labyrinth of forgotten lore and legend",
	"Chosen to wield the power of prophecy and fate",
	"Frozen in a world of fear and uncertainty",
	"Haunted by the horrors of the unknown and the supernatural",
	"Pursued by a powerful and relentless enemy",
	"Trapped in a world of mystery and intrigue",
	"Fighting for survival in a world of fantasy and science fiction",
	"Abandoned in a world of darkness and despair",
	"Lost in a sea of sorrow and suffering",
	"Chosen to wield the power of the elements and the forces of nature",
	"Frozen in a moment of desperation and hopelessness",
	"Trapped in a cycle of chaos and destruction",
	"Haunted by the ghosts of the past and the horrors of the present",
	"Pursued by a mysterious and malevolent force",
	"Fighting for survival in a post-apocalyptic world",
	"Trapped in a world of supernatural terror and horror",
	"Abandoned in a desolate and barren wasteland",
	"Lost in a labyrinth of forgotten knowledge and ancient lore",
	"Chosen to wield the power of prophecy and destiny",
	"Frozen in a world of ice and darkness",
	"Haunted by the horrors of the apocalypse and the forces of darkness",
	"Pursued by a powerful and ruthless enemy",
	"Trapped in a time loop of terror and despair",
	"Fighting for survival in a world of fantasy and horror",
	"Abandoned in a world of wonder and magic",
	"Lost in a sea of secrets and deception",
	"Chosen to wield the power of the gods and the forces of nature",
	"Frozen in a moment of madness and despair",
	"Trapped in a supernatural trap",
	"Haunted by the ghosts of war and violence",
	"Pursued by a monstrous and malevolent creature",
	"Fighting against an army of the undead and the forces of darkness",
	"Trapped in a cycle of chaos and destruction",
	"Abandoned in a desolate and barren land",
	"Lost in a labyrinth of forgotten lore and legend",
	"Chosen to wield the power of prophecy and fate",
	"Frozen in a world of fear and uncertainty",
	"Haunted by the horrors of the unknown and the supernatural",
	"Pursued by a powerful and relentless enemy",
	"Trapped in a world of mystery and intrigue",
	"Fighting for survival in a world of fantasy and science fiction",
	"Abandoned in a world of darkness and despair",
	"Lost in a sea of sorrow and suffering",
	"Chosen to wield the power of the elements and the forces of nature",
	"Frozen in a moment of desperation and hopelessness",
	"Trapped in a cycle of chaos and destruction",
	"Haunted by the ghosts of the past and the horrors of the present",
	"Pursued by a mysterious and malevolent force",
	"Fighting for survival in a post-apocalyptic world",
	"Trapped in a world of supernatural terror and horror",
	"Abandoned in a desolate and barren wasteland",
	"Lost in a labyrinth of forgotten knowledge and ancient lore",
	"Chosen to wield the power of prophecy and destiny",
	"Frozen in a world of ice and darkness",
	"Haunted by the horrors of the apocalypse and the forces of darkness",
	"Pursued by a powerful and ruthless enemy",
	"Trapped in a time loop of terror and despair",
	"Fighting for survival in a world of fantasy and horror",
	"Abandoned in a world of wonder and magic",
	"Lost in a sea of secrets and deception",
	"Chosen to wield the power of the gods and the forces of nature",
	"Frozen in a moment of madness and despair",
	"Trapped in a supernatural trap",
	"Haunted by the ghosts of war and violence",
	"Pursued by a monstrous and malevolent creature",
	"Fighting against an army of the undead and the forces of darkness",
	"Trapped in a cycle of chaos and destruction",
	"Abandoned in a desolate and barren land",
	"Lost in a labyrinth of forgotten lore and legend",
	"Chosen to wield the power of prophecy and fate",
	"Frozen in a world of fear and uncertainty",
	"Haunted by the horrors of the unknown and the supernatural",
	"Pursued by a powerful and relentless enemy",
	"Trapped in a world of mystery and intrigue",
	"Fighting for survival in a world of fantasy and science fiction",
	"Abandoned in a world of darkness and despair",
	"Lost in a sea of sorrow and suffering",
	"Chosen to wield the power of the elements and the forces of nature",
	"Frozen in a moment of desperation and hopelessness",
	"Trapped in a cycle of chaos and destruction",
	"Haunted by the ghosts of the past and the horrors of the present",
	"Pursued by a mysterious and malevolent force",
	"Fighting for survival in a post-apocalyptic world",
	"Trapped in a world of supernatural terror and horror",
	"Abandoned in a desolate and barren wasteland",
	"Lost in a labyrinth of forgotten knowledge and ancient lore",
	"Chosen to wield the power of prophecy and destiny",
	"Frozen in a world of ice and darkness",
	"Haunted by the horrors of the apocalypse and the forces of darkness",
	"Pursued by a powerful and ruthless enemy",
	"Trapped in a time loop of terror and despair",
	"Fighting for survival in a world of fantasy and horror",
	"Abandoned in a world of wonder and magic",
	"Lost in a sea of secrets and deception",
	"Chosen to wield the power of the gods and the forces of nature",
	"Frozen in a moment of madness and despair",
	"Trapped in a supernatural trap",
	"Haunted by the ghosts of war and violence",
	"Pursued by a monstrous and malevolent creature",
	"Fighting against an army of the undead and the forces of darkness",
	"Trapped in a cycle of chaos and destruction",
	"Abandoned in a desolate and barren land",
	"Lost in a labyrinth of forgotten lore and legend",
	"Chosen to wield the power of prophecy and fate",
	"Frozen in a world of fear and uncertainty",
	"Haunted by the horrors of the unknown and the supernatural",
	"Pursued by a powerful and relentless enemy",
	"Trapped in a world of mystery and intrigue",
	"Fighting for survival in a world of fantasy and science fiction",
	"Abandoned in a world of darkness and despair",
	"Lost in a sea of sorrow and suffering",
	"Chosen to wield the power of the elements and the forces of nature",
	"Frozen in a moment of desperation and hopelessness",
	"Trapped in a cycle of chaos and destruction",
	"Haunted by the ghosts of the past and the horrors of the present",
	"Pursued by a mysterious and malevolent force",
	"Fighting for survival in a post-apocalyptic world",
	"Trapped in a world of supernatural terror and horror",
	"Abandoned in a desolate and barren wasteland",
	"Lost in a labyrinth of forgotten knowledge and ancient lore",
	"Chosen to wield the power of prophecy and destiny",
	"Frozen in a world of ice and darkness",
	"Haunted by the horrors of the apocalypse and the forces of darkness",
	"Pursued by a powerful and ruthless enemy",
	"Trapped in a time loop of terror and despair",
	"Fighting for survival in a world of fantasy and horror",
	"Abandoned in a world of wonder and magic",
	"Lost in a sea of secrets and deception",
	"Chosen to wield the power of the gods and the forces of nature",
}

type MammouthConversationCreateDto struct {
	Message       string        `json:"message"`
	Model         string        `json:"model"`
	DefaultModels DefaultModels `json:"defaultModels"`
	AssistantID   interface{}   `json:"assistantId"`
	Attachments   []interface{} `json:"attachments"`
}

type DefaultModels struct {
	Text      string `json:"text"`
	Image     string `json:"image"`
	WebSearch string `json:"webSearch"`
}

type MammouthConversation struct {
	ID          int64       `json:"id"`
	AssistantID interface{} `json:"assistantId"`
	User        int64       `json:"user"`
	Title       string      `json:"title"`
	Model       string      `json:"model"`
	CreatedAt   time.Time   `json:"createdAt"`
	UpdatedAt   time.Time   `json:"updatedAt"`
	PublicID    interface{} `json:"publicId"`
}

func (c *MammouthConversation) SendMessage(client *MammouthClient, message string) (string, error) {
	form := new(bytes.Buffer)
	writer := multipart.NewWriter(form)
	formField, _ := writer.CreateFormField("model")
	formField.Write([]byte(c.Model))
	formField, _ = writer.CreateFormField("preprompt")
	formField.Write([]byte(""))
	formField, _ = writer.CreateFormField("messages")
	formField.Write([]byte(fmt.Sprintf(`{"content":"%s","imagesData":[],"documentsData":[]}`, message)))
	writer.Close()
	req, _ := http.NewRequest("POST", "https://mammouth.ai/api/models/llms", form)
	setHeaders(req)
	req.Header.Set("Content-Type", writer.FormDataContentType())
	resp, err := client.httpClient.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return "", err
	}
	body, _ := io.ReadAll(resp.Body)
	return string(body), nil
}

func (c *MammouthConversation) Delete(client *MammouthClient) error {
	data := strings.NewReader(fmt.Sprintf(`{"id":%d}`, c.ID))
	req, _ := http.NewRequest("POST", "https://mammouth.ai/api/chat/delete", data)
	setHeaders(req)
	resp, err := client.httpClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return err
	}

	return nil
}

func setHeaders(req *http.Request) {
	req.Header.Set("User-Agent", "Mozilla/5.0 (X11; Linux x86_64; rv:136.0) Gecko/20100101 Firefox/136.0")
	req.Header.Set("Accept", "application/json")
	req.Header.Set("Accept-Language", "en-US,en;q=0.5")
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Referer", "https://mammouth.ai/app/a/default")
	req.Header.Set("Origin", "https://mammouth.ai")
	req.Header.Set("DNT", "1")
	req.Header.Set("Sec-GPC", "1")
	req.Header.Set("Sec-Fetch-Dest", "empty")
	req.Header.Set("Sec-Fetch-Mode", "cors")
	req.Header.Set("Sec-Fetch-Site", "same-origin")
	req.Header.Set("Connection", "keep-alive")
	req.Header.Set("Priority", "u=0")
	req.Header.Set("TE", "trailers")
}

type MammouthClient struct {
	httpClient *http.Client
}

func (c *MammouthClient) CreateConversation(model, message string) (MammouthConversation, error) {
	dto := MammouthConversationCreateDto{
		Message:       message,
		Model:         model,
		DefaultModels: DefaultModels{Text: defaultTextModel, Image: "replicate-recraftai-recraft-v3", WebSearch: "openperplex-v1"},
		AssistantID:   nil,
		Attachments:   []interface{}{},
	}

	body, _ := json.Marshal(dto)

	req, _ := http.NewRequest("POST", "https://mammouth.ai/api/chat/create", bytes.NewBuffer(body))
	setHeaders(req)
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return MammouthConversation{}, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return MammouthConversation{}, err
	}

	var conversation MammouthConversation
	err = json.NewDecoder(resp.Body).Decode(&conversation)
	if err != nil {
		return MammouthConversation{}, err
	}

	return conversation, nil
}

func NewMammouthClient(token, gcpToken, language string) *MammouthClient {
	c := &http.Client{
		Timeout: time.Second * 10,
	}

	cookies := []*http.Cookie{
		{
			Name:     "auth_session",
			Value:    token,
			Domain:   "mammouth.ai",
			Path:     "/",
			Secure:   true,
			HttpOnly: false,
		},
		{
			Name:     "i18n_redirected",
			Value:    language,
			Domain:   "mammouth.ai",
			Path:     "/",
			Secure:   false,
			HttpOnly: false,
		},
		{
			Name:     "gcp_token",
			Value:    gcpToken,
			Domain:   "mammouth.ai",
			Path:     "/",
			Secure:   false,
			HttpOnly: false,
		},
	}

	url, _ := url.Parse("https://mammouth.ai/")

	c.Jar, _ = cookiejar.New(nil)
	c.Jar.SetCookies(url, cookies)

	return &MammouthClient{
		httpClient: c,
	}
}

func tokenCost(input string, tokenizer tokenizer.Codec, isInput bool) (int, float64) {
	ids, _, _ := tokenizer.Encode(input)
	if isInput {
		return len(ids), float64(len(ids)) * defaultPricePerTokenInput
	} else {
		return len(ids), float64(len(ids)) * defaultPricePerTokenOutput
	}
}

func verboseSleep(message string, t time.Duration) {
	log.Println("[#] sleeping for", t.String(), ":", message)
	time.Sleep(t)
}

func getTopic() string {
	return topics[rand.Intn(len(topics))]
}

func main() {
	enc, _ := tokenizer.Get(tokenizer.O200kBase)
	client := NewMammouthClient(
		"7ygl6hom3wkcnvwbw5zs2wq6b6xlbrbwoykwoemy",
		"39228_1746115549092_4593d53f5399202e40e489d6dff97c7433a4402d0abd9976e3264e4f6670b76c",
		"en",
	)
	sessCostInput := float64(0)
	sessCostOutput := float64(0)

	for {
		conversation, err := client.CreateConversation(
			"anthropic-claude-3-7-sonnet-latest",
			"Hello, I'd you to tell me a long...",
		)
		if err == nil {
			message := fmt.Sprintf("Hello, I'd like you to tell me an extremely long story about the topic: '%s'", getTopic())
			response, msgErr := conversation.SendMessage(client, message)
			{
				tokenCount, cost := tokenCost(message, enc, true)
				log.Printf("[#] input tokens: %d, cost: %.4f\n", tokenCount, cost)
				sessCostInput += cost
			}
			{
				tokenCount, cost := tokenCost(response, enc, true)
				log.Printf("[#] output tokens: %d, cost: %.4f\n", tokenCount, cost)
				sessCostOutput += cost
			}
			conversation.Delete(client)
			log.Printf("Total session I/O cost: %.4f€\n", sessCostInput+sessCostOutput)
			log.Println("--------------------------")

			if msgErr != nil {
				log.Println("[-] Error sending message:", msgErr)
				verboseSleep("waiting for next cycle (error)", time.Duration(rand.Intn(errMaxSleep-errMinSleep)+errMinSleep)*time.Millisecond)
				continue
			} else {
				verboseSleep("waiting for next cycle", time.Duration(rand.Intn(maxSleep-minSleep)+minSleep)*time.Millisecond)
			}
		} else {
			log.Println("[-] Error creating conversation:", err)
			verboseSleep("waiting for next cycle (error)", time.Duration(rand.Intn(errMaxSleep-errMinSleep)+errMinSleep)*time.Millisecond)
		}
	}
}
