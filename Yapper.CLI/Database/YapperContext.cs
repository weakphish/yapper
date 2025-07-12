using Microsoft.EntityFrameworkCore;

namespace Yapper.CLI.Database;

public class YapperContext : DbContext
{
    public DbSet<Models.Task> Tasks { get; set; }
    public DbSet<Models.Note> Notes { get; set; }

    protected override void OnConfiguring(DbContextOptionsBuilder optionsBuilder)
    {
        var dbPath = Path.Combine(Environment.GetFolderPath(Environment.SpecialFolder.UserProfile), ".yapper", "yapper.db");
        Directory.CreateDirectory(Path.GetDirectoryName(dbPath)!);
        optionsBuilder.UseSqlite($"Data Source={dbPath}");
    }

    protected override void OnModelCreating(ModelBuilder modelBuilder)
    {
        modelBuilder.Entity<Models.Task>(entity =>
        {
            entity.HasKey(t => t.Id);
            entity.Property(t => t.Title).IsRequired().HasMaxLength(200);
            entity.Property(t => t.Description).HasMaxLength(1000);
            entity.Property(t => t.Status).IsRequired().HasConversion<int>();
            entity.Property(t => t.CreatedAt).IsRequired();

            entity.HasOne(t => t.DependsOn)
                .WithMany(t => t.Dependents)
                .HasForeignKey(t => t.DependsOnId)
                .OnDelete(DeleteBehavior.SetNull);
        });

        modelBuilder.Entity<Models.Note>(entity =>
        {
            entity.HasKey(n => n.Id);
            entity.Property(n => n.Title).IsRequired().HasMaxLength(200);
            entity.Property(n => n.Content).HasMaxLength(1000);
            entity.Property(n => n.CreatedAt).IsRequired();
        });
    }
}