package cmd

import (
	"github.com/webbben/2d-game-engine/data/defs"
	"github.com/webbben/2d-game-engine/dialogv2"
)

var DRQ001FillOutMyself *defs.DialogResponse = &defs.DialogResponse{
	Text: "Ah, you can read! Lovely, that will fetch us a much better price for you. Very well, here you go.",
}

var DRQ001Questionaire1 *defs.DialogResponse = &defs.DialogResponse{
	Text: "Very good, this will only take a minute. Let's begin, shall we?",
	NextResponse: &defs.DialogResponse{
		Text: "Firstly, from what land do you hail? Are you a Phoenician pirate, or maybe a Greek mercenary? I suppose you could be a magical sort from Egypt, too." +
			"Or do you come from the forests of Germany or Gaul?",
		Replies: []defs.DialogReply{
			{
				Text: "Latin",
				NextResponse: &defs.DialogResponse{
					Text:         "Really? I would've never guessed! You sure do have an odd accent for a Roman. I suppose you'll make a good, disciplined soldier then. Nevermind, let's continue.",
					NextResponse: DRQ001Questionaire2,
				},
			},
			{
				Text: "Greek",
				NextResponse: &defs.DialogResponse{
					Text:         "Ah, a Greek I see. I suppose you may be handy with a spear, or perhaps you are knowledgeable in philosophy? Nevermind, let's continue.",
					NextResponse: DRQ001Questionaire2,
				},
			},
			{
				Text: "Phoenician",
				NextResponse: &defs.DialogResponse{
					Text:         "A Phoenician, eh? You've come a long way from Carthage or Sidon. I suppose you must be a sailor or merchant. Nevermind, let's continue.",
					NextResponse: DRQ001Questionaire2,
				},
			},
			{
				Text: "Egyptian",
				NextResponse: &defs.DialogResponse{
					Text:         "An Egyptian? You're a world away from your pyramids and crypts. Say, do you know the ways of Eastern magic? Oh nevermind, let's continue.",
					NextResponse: DRQ001Questionaire2,
				},
			},
			{
				Text: "Gallic",
				NextResponse: &defs.DialogResponse{
					Text:         "A Gaul? A long way from your hillfort, aren't you. I'll bet you are fiersome with a sword. Nevermind, let's continue",
					NextResponse: DRQ001Questionaire2,
				},
			},
			{
				Text: "Germanic",
				NextResponse: &defs.DialogResponse{
					Text:         "A German? You're a long way from the forests of Germania, aren't you. I have no doubt you are terrifying on the battlefield though. Ah nevermind, let's continue.",
					NextResponse: DRQ001Questionaire2,
				},
			},
		},
	},
}

var DRQ001Questionaire2 *defs.DialogResponse = &defs.DialogResponse{
	Text: "Next, I will ask you how you make your living. What is your trade?",
	Replies: []defs.DialogReply{
		{
			Text: "Soldier",
			NextResponse: &defs.DialogResponse{
				Text:         "Interesting, I suppose that means you'll be quite handy with a sword or spear. You'll make a great legionary.",
				NextResponse: DRQ001Questionaire3,
			},
		},
		{
			Text: "Hunter",
			NextResponse: &defs.DialogResponse{
				Text:         "Interesting, you must be quite skilled with a bow, and familiar with the forest. You could make a good scout with the Exploratores.",
				NextResponse: DRQ001Questionaire3,
			},
		},
		{
			Text: "Barbarian",
			NextResponse: &defs.DialogResponse{
				Text:         "Well, not the most sophisticated sort, but I sure wouldn't want to meet you on the battlefield. You'll do well as an Auxiliary.",
				NextResponse: DRQ001Questionaire3,
			},
		},
		{
			Text: "Scholar",
			NextResponse: &defs.DialogResponse{
				Text:         "A bookish sort? Well I hope you've read up on military strategy then. For your own sake, we'll try to find a spot for you that isn't on the front line...",
				NextResponse: DRQ001Questionaire3,
			},
		},
		{
			Text: "This n' That",
			NextResponse: &defs.DialogResponse{
				Text:         "Interesting... You better watch yourself, because the Centurions are not afraid to dish out punishment for mischief. Maybe you would do well as a spy with the Speculatores.",
				NextResponse: DRQ001Questionaire3,
			},
		},
	},
}

var DRQ001Questionaire3 *defs.DialogResponse = &defs.DialogResponse{
	Text: "Next is the personality section. I will ask you a few questions to ascertain your way of thinking and your sense of morality.",
	NextResponse: &defs.DialogResponse{
		Text: "You are shopping for produce in a busy marketplace, when an elderly man rudely pushes past you. To add insult to injury, he barks at you, " +
			"'watch where you're going, fool!'. How do you handle this situation?",
		Replies: []defs.DialogReply{
			{
				Text: "Turn and move on with my business. It's no matter to me.",
				NextResponse: &defs.DialogResponse{
					Text:         "That's very noble of you. Jupiter smiles on such conduct.",
					NextResponse: DRQ001Questionaire4,
				},
			},
			{
				Text: "Draw my mace and crush his skull in for his insolence!",
				NextResponse: &defs.DialogResponse{
					Text:         "Gods! What a violent man you are... Ahem, let's move on.",
					NextResponse: DRQ001Questionaire4,
				},
			},
			{
				Text: "Apologize, and while patting him on the shoulder, stealthily swipe his beloved amulet from his neck.",
				NextResponse: &defs.DialogResponse{
					Text:         "What kind of underhanded theivery-- hm, well, I guess he DID have it coming. Nevermind, moving on.",
					NextResponse: DRQ001Questionaire4,
				},
			},
			{
				Text: "Apologize, and offer him one of my apples - which I've secretly coated with rat poison!",
				NextResponse: &defs.DialogResponse{
					Text:         "Oh my, are you seriously saying... Surely this is some kind of joke? Nevermind, let's move on.",
					NextResponse: DRQ001Questionaire4,
				},
			},
		},
	},
}

var DRQ001Questionaire4 *defs.DialogResponse = &defs.DialogResponse{
	Text: "You are praying at the temple of Apollo one quiet evening when you hear the faint sounds of a scuffle occuring. The sounds seem to be coming from the treasury room. " +
		"Creeping around the corner, you find the door unlocked and there are two cloaked men threatening the priest with a dagger! You quickly realize these men are thieves, " +
		"robbing the sacred temple of Apollo! What do you do?",
	Replies: []defs.DialogReply{
		{
			Text: "Run away! I'm not risking my life, this is none of my business!",
			NextResponse: &defs.DialogResponse{
				Text:         "How cowardly of you! Well, now that I think of it, I'm not much of a fighting man myself... Still!",
				NextResponse: DRQ001Questionaire5,
			},
		},
		{
			Text: "Invoke the power of the gods before charging upon them with my sword! Nobody defiles a sanctuary of Apollo!",
			NextResponse: &defs.DialogResponse{
				Text:         "That is the answer that a zealous champion of Mars would give. Well done!",
				NextResponse: DRQ001Questionaire5,
			},
		},
		{
			Text: "Wait for them to deal with the priest, and slip in past them to grab a trinket for myself.",
			NextResponse: &defs.DialogResponse{
				Text:         "What treachery do you speak of? You must be quite the lowlife scoundrel yourself.",
				NextResponse: DRQ001Questionaire5,
			},
		},
		{
			Text: "What have the gods done for me? I'd just carry on minding my own business, and let Apollo sort it out.",
			NextResponse: &defs.DialogResponse{
				Text:         "A cynical answer? Perhaps if you put more faith in the gods, you wouldn't be where you are today.",
				NextResponse: DRQ001Questionaire5,
			},
		},
	},
}

var DRQ001Questionaire5 *defs.DialogResponse = &defs.DialogResponse{
	Text: "One more scenario for you: You're a town guard on patrol, and you turn a corner to find a man climbing out of a window with a sack slung over his shoulder. " +
		"The sack jingles with the sound of gold coins, and it's clear you've just caught a robber red handed. However, the man pleads for his life and tells you he is just trying to " +
		"support his impoverished, starving family. What do you do?",
	Replies: []defs.DialogReply{
		{
			Text: "Immediately place him under arrest and bring him to the gallows. No man breaks the law on my watch.",
			NextResponse: &defs.DialogResponse{
				Text:         "Such is the word of the law, which is your duty to carry out. Very good.",
				NextResponse: DRQ001QuestionaireEnd,
			},
		},
		{
			Text: "Let the poor man go this time; The rich have plenty of gold, and his family needs it more.",
			NextResponse: &defs.DialogResponse{
				Text:         "Hm... You might think this is the honorable way, but what about the rich? Who is going to look out for them? No matter, let's continue.",
				NextResponse: DRQ001QuestionaireEnd,
			},
		},
		{
			Text: "Shake him down for the loot, threatening to kill him if he doesn't hand it over. What is he gonna do, call the guards?",
			NextResponse: &defs.DialogResponse{
				Text:         "By Jupiter! You really are some sort of villain, aren't you?",
				NextResponse: DRQ001QuestionaireEnd,
			},
		},
		{
			Text: "Order him to return the gold, and give him money from your own pocket to buy a loaf of bread. Nobody deserves to go hungry.",
			NextResponse: &defs.DialogResponse{
				Text:         "Well, that is very kind and pious of you. But how can you be sure he's telling the truth? No matter, let's continue.",
				NextResponse: DRQ001QuestionaireEnd,
			},
		},
	},
}

var DRQ001QuestionaireEnd *defs.DialogResponse = &defs.DialogResponse{
	Text:         "... And, I believe that was the last question. Good, I think we're done here. Ah yes, perhaps I should explain some things to you now.",
	NextResponse: DRQ001Conscripted,
}

var DRQ001CharCreationEnd *defs.DialogResponse = &defs.DialogResponse{
	Text:         "Ah yes, you're finished? Let me look it over really quick... Hm... Interesting. Good, I think we're done here. Ah, perhaps I should explain some things to you now.",
	NextResponse: DRQ001Conscripted,
}

var DRQ001Conscripted *defs.DialogResponse = &defs.DialogResponse{
	Text: "You are hereby a slave and are property of the Roman legion - yes, it says here in your papers that you will be conscripted into a Legion camped near the " +
		"frontier in Germania.",
	NextResponse: &defs.DialogResponse{
		Text: "... Hm? Why the long face? Oh, the whole 'slavery' thing? I don't know what you expected. You should be grateful that I haven't sent you to the iron mines. " +
			"Those are hideous places, and your odds of survival are much better as a soldier. However, I must warn you that the campaigns in Germania have been particularly brutal. " +
			"Those tribal warriors of Germania can inspire terror in the hearts of the bravest of men, so keep your wits about you.",
		NextResponse: &defs.DialogResponse{
			Text: "Hm, let's see here... ah yes, and I'm required to tell you that heroic deeds and outstanding gallantry in the service of Caesar may make you eligible for emancipation, " +
				"failure to comply with the orders of your superiors can be punished by torturous death or crucifixion, yadda yadda yadda... And... looks like we're finished with the fineprint.",
			NextResponse: &defs.DialogResponse{
				Text: "Great, well I think we're done here now, so I'll hand you a copy of the paperwork and the officer at the door will take you to your bunk. I believe you set out first thing in " +
					"the morning. You have a very long journey ahead of you still. Good luck, and may the gods have mercy on you, " + dialogv2.VarPlayerName + ".",
				Effects: []defs.DialogEffect{
					dialogv2.EventEffect{
						Event: defs.Event{
							Type: Q001CharacterCreationComplete,
						},
					},
				},
			},
		},
	},
}
