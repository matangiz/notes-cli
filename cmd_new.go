package notes

import (
	"fmt"
	"github.com/pkg/errors"
	"gopkg.in/alecthomas/kingpin.v2"
	"io/ioutil"
	"os"
)

// NewCmd represents `notes new` command. Each public fields represent options of the command
type NewCmd struct {
	cli    *kingpin.CmdClause
	Config *Config
	// Category is a category name of the new note. This must be a name allowed for directory name
	Category string
	// Filename is a file name of the new note
	Filename string
	// Tags is a comma-separated string of tags of the new note
	Tags string
	// NoInline is a flag equivalent to --no-inline-input
	NoInline bool
}

func (cmd *NewCmd) defineCLI(app *kingpin.Application) {
	cmd.cli = app.Command("new", "Create a new note")
	cmd.cli.Arg("category", "Category of memo").Required().StringVar(&cmd.Category)
	cmd.cli.Arg("filename", "Name of memo").Required().StringVar(&cmd.Filename)
	cmd.cli.Arg("tags", "Comma-separated tags of memo").StringVar(&cmd.Tags)
	cmd.cli.Flag("no-inline-input", "Does not request inline input even if no editor is set").BoolVar(&cmd.NoInline)
}

func (cmd *NewCmd) matchesCmdline(cmdline string) bool {
	return cmd.cli.FullCommand() == cmdline
}

func (cmd *NewCmd) fallbackInput(note *Note) error {
	fmt.Fprintln(os.Stderr, "Input notes inline (Send EOF by Ctrl+D to stop):")
	b, err := ioutil.ReadAll(os.Stdin)
	if err != nil {
		return errors.Wrap(err, "Cannot read from stdin")
	}

	f, err := os.OpenFile(note.FilePath(), os.O_APPEND|os.O_WRONLY, 0644)
	if err != nil {
		return errors.Wrap(err, "Cannot open note file")
	}
	if _, err := f.Write(b); err != nil {
		return errors.Wrap(err, "Cannot write to note file")
	}

	fmt.Fprintln(os.Stderr)
	fmt.Println(note.FilePath())

	return nil
}

// Do runs `notes new` command and returns an error if occurs
func (cmd *NewCmd) Do() error {
	git := NewGit(cmd.Config)

	note, err := NewNote(cmd.Category, cmd.Tags, cmd.Filename, "", cmd.Config)
	if err != nil {
		return err
	}

	if err := note.Create(); err != nil {
		return err
	}

	if git != nil {
		if err := git.Init(); err != nil {
			return err
		}
	}

	if err := note.Open(); err != nil {
		if !cmd.NoInline {
			fmt.Fprintf(os.Stderr, "Note: %s\n", err)
		}
		if !cmd.NoInline {
			return cmd.fallbackInput(note)
		}
		// Final fallback is only showing the path to the note. Then users can open it by themselves.
		fmt.Println(note.FilePath())
	}

	return nil
}
