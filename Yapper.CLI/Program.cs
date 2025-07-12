using Spectre.Console;
using Spectre.Console.Cli;
using Yapper.CLI.Commands;

namespace Yapper.CLI;

class Program
{
    static async Task<int> Main(string[] args)
    {
        try
        {
            var app = new CommandApp();
            app.Configure(config =>
            {
                config.SetApplicationName("yapper");
                config.SetApplicationVersion("1.0.0");
                
                config.AddBranch("task", task =>
                {
                    task.SetDescription("Manage tasks");
                    task.AddCommand<AddTaskCommand>("add")
                        .WithDescription("Add a new task");
                    
                    task.AddCommand<ListTasksCommand>("list")
                        .WithDescription("List all tasks")
                        .WithAlias("ls");
                });
            });
            
            return await app.RunAsync(args);
        }
        catch (Exception ex)
        {
            AnsiConsole.WriteException(ex);
            return 1;
        }
    }
}
