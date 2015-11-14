CREATE DATABASE IF NOT EXISTS sj;
CREATE TABLE Series (
	ID int AUTO_INCREMENT PRIMARY KEY,
	Title varchar(250),
	Image varchar(500)
);
CREATE TABLE EpisodesResource (
	ID int AUTO_INCREMENT PRIMARY KEY,
	Series_ID int,
	Name varchar(250),
	URL varchar(500),
	FOREIGN KEY(Series_ID) REFERENCES Series(ID)
);
CREATE TABLE Episodes (
	ID int AUTO_INCREMENT PRIMARY KEY,
	Series_ID int,
	Title varchar(500),
	Session int,
	Episode int,
	FOREIGN KEY(Series_ID) REFERENCES Series(ID)
);
CREATE TABLE User(
	ID int AUTO_INCREMENT PRIMARY KEY,
	Name varchar(500),
	Password varchar(136)
);
CREATE TABLE SeriesList (
	User_ID int NOT NULL,
	Series_ID int NOT NULL,
	PRIMARY KEY (User_ID, Series_ID)
);
CREATE TABLE LastWatched (
	User_ID int NOT NULL,
	Series_ID int NOT NULL,
	Session int,
	Episode int,
	PRIMARY KEY (User_ID, Series_ID)
)
