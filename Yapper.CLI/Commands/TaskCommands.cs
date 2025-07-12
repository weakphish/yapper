using Spectre.Console;
using Spectre.Console.Cli;
using System.ComponentModel;
using Microsoft.EntityFrameworkCore;
using Yapper.CLI.Database;
using TaskStatus = Yapper.CLI.Models.TaskStatus;

namespace Yapper.CLI.Commands;

public class AddTaskCommand : AsyncCommand<AddTaskCommand.Settings>
{
    public class Settings : CommandSettings
    {
        [CommandArgument(0, "<TITLE>")]
        [Description("The title of the task")]
        public string Title { get; set; } = string.Empty;

        [CommandOption("-d|--description")]
        [Description("The description of the task")]
        public string? Description { get; set; }

        [CommandOption("--depends-on")]
        [Description("ID of the task this task depends on")]
        public int? DependsOnId { get; set; }
    }

    public override async System.Threading.Tasks.Task<int> ExecuteAsync(CommandContext context, Settings settings)
    {
        try
        {
            using var db = new YapperContext();
            db.Database.EnsureCreated();

            // Find dependency if specified
            Yapper.CLI.Models.Task? dependsOn = null;
            if (settings.DependsOnId.HasValue)
            {
                dependsOn = await db.Tasks.FindAsync(settings.DependsOnId.Value);
                if (dependsOn == null)
                {
                    AnsiConsole.MarkupLine($"[red]Task with ID {settings.DependsOnId.Value} not found[/]");
                    return 1;
                }
            }

            var task = new Yapper.CLI.Models.Task
            {
                Title = settings.Title,
                Description = settings.Description ?? string.Empty,
                Status = TaskStatus.Todo,
                CreatedAt = DateTime.UtcNow,
                DependsOn = dependsOn,
                DependsOnId = dependsOn?.Id
            };

            db.Tasks.Add(task);
            await db.SaveChangesAsync();
            
            AnsiConsole.MarkupLine($"[green]✓[/] Task '{task.Title}' added successfully!");
            return 0;
        }
        catch (Exception ex)
        {
            AnsiConsole.WriteException(ex);
            return 1;
        }
    }
}

public class ListTasksCommand : AsyncCommand
{
    public override async System.Threading.Tasks.Task<int> ExecuteAsync(CommandContext context)
    {
        try
        {
            using var db = new YapperContext();
            db.Database.EnsureCreated();
            
            var tasks = await db.Tasks.Include(t => t.DependsOn).ToListAsync();
            
            if (tasks.Count == 0)
            {
                AnsiConsole.MarkupLine("[yellow]No tasks found. Use 'yapper task add <title>' to create your first task.[/]");
                return 0;
            }

            var table = new Table();
            table.AddColumn("ID");
            table.AddColumn("Title");
            table.AddColumn("Status");
            table.AddColumn("Created");
            table.AddColumn("Depends On");

            foreach (var task in tasks.OrderBy(t => t.Id))
            {
                var statusColor = task.Status switch
                {
                    TaskStatus.Todo => "yellow",
                    TaskStatus.InProgress => "blue", 
                    TaskStatus.Completed => "green",
                    _ => "white"
                };

                table.AddRow(
                    task.Id.ToString(),
                    task.Title,
                    $"[{statusColor}]{task.Status}[/]",
                    task.CreatedAt.ToString("yyyy-MM-dd"),
                    task.DependsOn?.Title ?? "-"
                );
            }

            AnsiConsole.Write(table);
            return 0;
        }
        catch (Exception ex)
        {
            AnsiConsole.WriteException(ex);
            return 1;
        }
    }
}