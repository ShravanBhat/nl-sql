# Go SQL Agent

This is a tool that helps you talk to your database using plain English. You do not need to be an expert in SQL to use it. It connects to your database and uses artificial intelligence to translate your questions into database queries.

## How it works

The application has a simple web interface. You start by entering your database connection details. It currently supports PostgreSQL and MySQL.

Once connected, you can type a question like "Show me the users who signed up yesterday" or "Count the number of orders." The tool uses Google's Gemini AI to understand your question and write the SQL code for you.

You will always see the generated SQL code before it runs. This allows you to double check it. If it looks correct, you can click a button to execute it and see the results in a table.

## How to set it up

You need to have Go installed on your computer to run this program. You also need an API key for Google Gemini.

1.  **Get the code**
    Download or clone this project to your computer.

2.  **Set up your API key**
    Create a new file in the project folder named `.env`. Inside this file, add your Gemini API key like this:
    
    GEMINI_API_KEY=your_api_key_here

3.  **Run the application**
    Open your terminal or command prompt in the project folder and run this command:
    
    go run main.go

4.  **Open in your browser**
    Once the server starts, open your web browser and go to `http://localhost:8080`.

## Technologies used

This project is built with Go for the backend server. The frontend uses HTML and Tailwind CSS for styling. It uses the Gemini API for the intelligence part.

## High Level Design

The system is built with three main parts that work together.

1.  **The User Interface**
    This is the website you see in your browser. It takes your input and shows you the results. It talks to the backend server.

2.  **The Backend Server**
    This is the main program running on your computer. It acts as the middleman. It receives your questions, talks to the AI, and runs commands on your database.

3.  **External Services**
    We connect to two outside things. First is your database (like PostgreSQL or MySQL) where your data lives. Second is Google Gemini, which acts as the brain to translate English into SQL.

## Low Level Design

Organized the code into specific folders to keep it clean and easy to understand.

*   **main.go**: This is the starting point. It turns on the server and loads your settings.
*   **api folder**: This contains the code that handles requests from the website. For example, when you click "Connect", a specific function in this folder does the work.
*   **core/db folder**: This handles all the messy details of talking to different databases. It makes sure that asking for data looks the same to the rest of the app, whether you use Postgres or MySQL.
*   **core/ai folder**: This manages the conversation with Google Gemini. It prepares your question and the database structure so the AI can understand it.

## Design Patterns Used

We used standard coding patterns to solve common problems. Here is what we used and why.

### Factory Pattern
We use a "Factory" to create the database connections. When you choose a database type in the dropdown, the Factory picks the correct code to run.

**Why we used it:** This makes it very easy to add new databases in the future. If we want to add SQLite support later, we just add a new piece to the Factory without breaking the existing code.

### Adapter Pattern
We created a standard interface (a set of rules) for how to talk to a database. Then we wrote specific "Adapters" for Postgres and MySQL that follow these rules.

**Why we used it:** Different databases speak slightly different languages. The Adapter makes them all look the same to our application. This means the main application logic does not need to worry about which specific database you are using.
