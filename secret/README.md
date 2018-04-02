## Secret service ##

Secret service provide convenient way of handling credentials.
Service uses [credential config](./../cred/config.go) to store various provided credentials.


## Credentials retrieval

Service supports the following form of **secret** to retrieve credentials:

1. URL i.e. mem://secret/localhost.json
2. Relative path i.e. localhost.json, in this case based directory will be used to lookup credential resource
3. Short name  i.e. localhost, in this case based directory will be used to lookup credential resource and .json ext will be added.
4. Inline certificates or secrets that does not represent resource location are placed into cred.Config.Data


**Base directory** can be file or URL, if empty '$HOME/.secret/' is used 


```go

    service := New(baseDirectory, false) 
    var secret = "localhost"
    service.GetCredentials(secret)


```

## Credentials generation

Service allows for interactive credential generation, in this scenario services asks 
a user for username and password for supplied secret. 
Optionally private key path can be supplied for pub key based auth.

```go
    
    privateKeyPath := ""//optional path
    secret := "localhost"
    service := New(baseDirectory, false)
    location, err := service.Create(secret, privateKeyPath)
    
      
```

The following example uses service in the interactive mode, which will try to lookup credential
for supplied key, if failed it will ask a user for credential in terminal.


```go
  
    service := New(baseDirectory, true)
    config, err := service.GetOrCreate("xxxx")
      
```


## Secret expansion

Very common case for the application it to take encrypted credential to used wither username or password.
For example while running terminal command we may need to provide super user password and sometimes other secret, 
in one command that we do not want to reveal to final user.


Take the following code as example:

```go
        

    service := New(baseDirectory, true)
    secrets := NewSecrets()
    {//password expansion
        secrets["mysql"] = "~/.secret/mysql.json"
        input := "docker run --name db1 -e MYSQL_ROOT_PASSWORD=${mysql.password} -d mysql:tag"
   	    expaned, err := service.Expand(input, secrets)
   	}

   	{//username and password expansion
        secrets["pg"] = "~/.secret/pg.json"
        input := "docker run --name some-postgres -e POSTGRES_PASSWORD=${pg.password} -e POSTGRES_USER=${pg.username} -d postgres"
        expaned, err := service.Expand(input, secrets)
    }
  
   
    

```

Secrets represents credentials secret map defined as `type Secrets map[SecretKey]Secret`

Here are some possible combination of secret map pairs.

```json
 {
    "git": "${env.HOME}/.secret/git.json",
    "github.com": "${env.HOME}/.secret/github.json",
    "github.private.com": "${env.HOME}/.secret/github-private.json",
    "**replace**": "${env.HOME}/.secret/git.json"
 }
```

The secret key can be static or dynamic. The first type in input/command is enclosed with either '*' or '#', the later is not.

In the command corresponding dynamic key can be enclosed with the following
1) '${secretKey.password}' for password expansion  i.e.  command: **${git.password}**  will expand to password from  git secret key
2) '**' for password expansion  i.e.  command: `**git**`will expand to password from  git secret key
3) '${secretKey.username}'  for username expansion  i.e.  command: **${git.username}** will expand to username from  git secret key
4) '##' for username expansion  i.e.  command: `##git##` will expand to username from  git secret key
