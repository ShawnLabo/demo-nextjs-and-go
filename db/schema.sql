CREATE TABLE Accounts (
  AccountId    STRING(36)  NOT NULL,
  ApiToken     STRING(MAX) NOT NULL,
  Email        STRING(256) NOT NULL,
  Name         STRING(MAX) NOT NULL,
  LastAccessed TIMESTAMP,
) PRIMARY KEY (AccountId);

CREATE UNIQUE INDEX ApiTokenIndex ON Accounts (ApiToken);
CREATE UNIQUE INDEX EmailIndex ON Accounts (Email);
