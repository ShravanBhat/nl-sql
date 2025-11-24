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
