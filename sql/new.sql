CREATE DATABASE IF NOT EXISTS sj ;
CREATE TABLE Series (
	ID INT AUTO_INCREMENT PRIMARY KEY,
	Title varchar(250),
	Image varchar(500)
);
CREATE TABLE EpisodesResource (
	ID INT AUTO_INCREMENT PRIMARY KEY,
	Series_ID int,
	Name varchar(250),
	URL varchar(500),
	FOREIGN KEY(Series_ID) REFERENCES Series(ID)
);
CREATE TABLE Episodes (
	ID INT AUTO_INCREMENT PRIMARY KEY,
	Series_ID int,
	Title varchar(500),
	Session int,
	Episode int,
	FOREIGN KEY(Series_ID) REFERENCES Series(ID)
)
