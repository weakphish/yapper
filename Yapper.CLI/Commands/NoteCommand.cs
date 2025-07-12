using System.ComponentModel;
using Spectre.Console;
using Spectre.Console.Cli;
using Yapper.CLI.Database;

public class AddNoteCommand : AsyncCommand<AddNoteCommand.Settings>
{


    public class Settings : CommandSettings
    {
        [CommandArgument(0, "[TITLE]")]
        [Description("The title of the note")]
        public string? Title { get; set; }

    }
    public override async Task<int> ExecuteAsync(CommandContext context, Settings settings)
    {
        try
        {
            using var db = new YapperContext();
            db.Database.EnsureCreated();

            // If no title provided, show interactive form
            string title = settings.Title ?? AnsiConsole.Ask<string>("What's the [green]title[/] of your note?");
            
            string content = AnsiConsole.Prompt(
                new TextPrompt<string>("Enter the [yellow]content[/] of your note:")
                    .AllowEmpty()
                    .DefaultValue(string.Empty));

            var note = new Yapper.CLI.Models.Note
            {
                Title = title,
                Content = content,
                CreatedAt = DateTime.UtcNow
            };

            db.Notes.Add(note);
            await db.SaveChangesAsync();
            
            AnsiConsole.MarkupLine($"[green]✓[/] Note '{note.Title}' added successfully!");
            return 0;
        }
        catch (Exception ex)
        {
            AnsiConsole.WriteException(ex);
            return 1;
        }
    }
}