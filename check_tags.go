package publiccode

// checkTag tells whether the tag is a valid tag or not and returns it.
func (p *parser) checkTag(key, tag string) (string, error) {
	for _, t := range tagList {
		if t == tag {
			return tag, nil
		}
	}
	return tag, newErrorInvalidValue(key, "unknown tag: %s", tag)
}

// A tagList represents a list of the possible tags.
var tagList = []string{
	// International
	"3dgraphics",    //application for viewing, creating, or processing 3-d graphics
	"accessibility", //accessibility
	"accounting",    //accounting software
	"amusement",     //a simple amusement
	"archiving",     //a tool to archive/backup data
	"art",           //software to teach arts
	"artificialintelligence", //artificial intelligence software
	"backend",                //software not meant for end users
	"calculator",             //a calculator
	"calendar",               //calendar application
	"chat",                   //a chat client
	"classroom-management",   //classroom management software
	"clock",                  //a clock application/applet
	"cms",                    //a content management system
	"compression",            //a tool to manage compressed data/archives
	"construction",           // a tool for constructions
	"contactmanagement",      //e.g. an address book
	"database",               //application to manage a database
	"debugger",               //a tool to debug applications
	"dictionary",             //a dictionary
	"documentation",          //help or documentation
	"electronics",            //electronics software, e.g. a circuit designer
	"email",                  //email application
	"emulator",               //emulator of another platform, such as a dos emulator
	"engineering",            //engineering software, e.g. cad programs
	"filemanager",            //a file manager
	"filetransfer",           //tools like ftp or p2p programs
	"finance",                //application to manage your finance
	"flowchart",              //a flowchart application
	"guidesigner",            //a gui designer application
	"identity",               //identity management
	"instantmessaging",       //an instant messaging client
	"languages",              //software to learn foreign languages
	"library",                //a library software
	"medicalsoftware",        //medical software
	"monitor",                //monitor application/applet that monitors some resource or activity
	"museum",                 //museum software
	"music",                  //musical software
	"news",                   //software to manage and publish news
	"ocr",                    //optical character recognition application
	"parallelcomputing",      //parallel computing software
	"photography",            //camera tools, etc.
	"presentation",           //presentation software
	"printing",               //a tool to manage printers
	"procurement",            //software for procurement
	"projectmanagement",      //project management application
	"publishing",             //desktop publishing applications and color management tools
	"rastergraphics",         //application for viewing, creating, or processing raster (bitmap) graphics
	"remoteaccess",           //a tool to remotely manage your pc
	"revisioncontrol",        //applications like cvs or subversion
	"robotics",               //robotics software
	"scanning",               //tool to scan a file/text
	"security",               //a security tool
	"sports",                 //sports software
	"spreadsheet",            //a spreadsheet
	"telephony",              //telephony via pc
	"terminalemulator",       //a terminal emulator application
	"texteditor",             //a text editor
	"texttools",              //a text tool utility
	"translation",            //a translation tool
	"vectorgraphics",         //application for viewing, creating, or processing vector graphics
	"videoconference",        //video conference software
	"viewer",                 //tool to view e.g. a graphic or pdf file
	"webbrowser",             //a web browser
	"whistleblowing",         //software for whistleblowing / anticorruption
	"wordprocessor",          //a word processor
	// Italy.
	"it-portale-trasparenza",
}
