using System.ComponentModel.DataAnnotations;
using System.ComponentModel.DataAnnotations.Schema;

namespace Yapper.CLI.Models;



public class Note
{
    [Key]
    public int Id { get; set; }

    [Required]
    [MaxLength(200)]
    public string Title { get; set; } = string.Empty;

    [MaxLength(1000)]
    public string Content { get; set; } = string.Empty;


    public DateTime CreatedAt { get; set; } = DateTime.UtcNow;

    public DateTime? StartedAt { get; set; }

    public DateTime? CompletedAt { get; set; }

    [ForeignKey("DependsOn")]


    public virtual ICollection<Task> RelatedTasks { get; set; } = new List<Task>();
    public virtual ICollection<Note> RelatedNotes { get; set; } = new List<Note>();

    public static Note CreateNew(string title, string content)
    {
        return new Note
        {
            Title = title,
            Content = content,
            CreatedAt = DateTime.UtcNow
        };
    }
}