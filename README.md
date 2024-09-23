
# Gopher CLI Manager 🐹💻

Gopher CLI Manager is a terminal user interface (TUI) application for managing Command-Line Interface (CLI) tools. It allows you to easily add, update, search, and remove CLI tools stored in a SQLite database. Built with Go, this project leverages `bubbletea` for the interactive UI and `sqlite3` for the database management. 🚀

## Features ✨
- 📋 **View all CLI tools**: List all stored CLI tools.
- 🔍 **Search for CLIs**: Search for CLI tools by name or description.
- ➕ **Add new CLIs**: Easily add a new CLI with name, description, and path.
- ✏️ **Edit existing CLIs**: Update the information for an existing CLI.
- 🗑️ **Delete CLIs**: Remove CLI tools from the database.

## Screenshot 📸
![App Screenshot](https://via.placeholder.com/800x400.png?text=App+Screenshot)  
_Add a screenshot of your TUI application here_

## Installation 🛠️
1. Clone the repository:
   ```bash
   git clone https://github.com/yourusername/gopher-cli-manager.git
   ```
2. Navigate into the project directory:
   ```bash
   cd gopher-cli-manager
   ```
3. Install dependencies:
   ```bash
   go mod tidy
   ```
4. Build the project:
   ```bash
   make build
   ```

## Usage 🚀
After building the project, you can run it using:
```bash
make run
```

### CLI Management Options
- Press `v` to view all CLI tools.
- Press `s` to search for CLI tools.
- Press `a` to add a new CLI tool.
- Press `q` to quit the program.

## Database Schema 🗄️
The SQLite database contains a single table called `cli` with the following columns:
- `id`: Unique identifier (INTEGER, PRIMARY KEY, AUTOINCREMENT)
- `name`: The name of the CLI tool (TEXT)
- `description`: A brief description of the CLI tool (TEXT)
- `path`: The path to the CLI tool (TEXT)

## Contributing 🤝
Feel free to open issues or submit pull requests! Contributions are welcome.

## License 📄
This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.

## Future Enhancements 🔮
- Add more fields to the CLI database (e.g., version, category).
- Implement a command-line argument interface for advanced usage.
- Support for additional database options (e.g., PostgreSQL, MySQL).

## Credits 💡
- [Charm](https://github.com/charmbracelet/bubbletea) for the `bubbletea` framework.
- [Mattn](https://github.com/mattn/go-sqlite3) for the SQLite Go driver.

![Project Logo](https://via.placeholder.com/400x200.png?text=Project+Logo)  
