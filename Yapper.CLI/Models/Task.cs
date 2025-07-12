using System.ComponentModel.DataAnnotations;
using System.ComponentModel.DataAnnotations.Schema;

namespace Yapper.CLI.Models;

public enum TaskStatus
{
    Todo = 0,
    InProgress = 1,
    Completed = 2
}

public class Task
{
    [Key]
    public int Id { get; set; }
    
    [Required]
    [MaxLength(200)]
    public string Title { get; set; } = string.Empty;
    
    [MaxLength(1000)]
    public string Description { get; set; } = string.Empty;
    
    public TaskStatus Status { get; set; } = TaskStatus.Todo;
    
    public DateTime CreatedAt { get; set; } = DateTime.UtcNow;
    
    public DateTime? StartedAt { get; set; }
    
    public DateTime? CompletedAt { get; set; }
    
    [ForeignKey("DependsOn")]
    public int? DependsOnId { get; set; }
    
    public virtual Task? DependsOn { get; set; }
    
    public virtual ICollection<Task> Dependents { get; set; } = new List<Task>();

    public static Task CreateNew(string title, string description)
    {
        return new Task
        {
            Title = title,
            Description = description,
            Status = TaskStatus.Todo,
            CreatedAt = DateTime.UtcNow
        };
    }
}