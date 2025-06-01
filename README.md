# go-flutter-parcial2_api

API en Go usando Gin, GORM y autenticación JWT, conectada a Postgres y dockerizada.

## Arquitectura
- **controllers/**: Lógica de negocio y controladores HTTP
- **models/**: Modelos de datos (GORM)
- **routes/**: Definición de rutas y agrupación
- **config/**: Configuración de base de datos y variables de entorno
- **middlewares/**: Middlewares personalizados (ej: autenticación JWT)
- **utils/**: Utilidades auxiliares

## Endpoints principales
- `POST /auth/register` — Registro de usuario
- `POST /auth/login` — Login y obtención de JWT
- `GET /protected/profile` — Ruta protegida (requiere JWT)

## Variables de entorno
Configura las variables en un archivo `.env` o usa las del sistema:
```
DB_HOST=localhost
DB_PORT=5432
DB_USER=postgres
DB_PASSWORD=postgres
DB_NAME=goauth
JWT_SECRET=supersecretkey
```

## Levantar con Docker Compose

1. Construye y levanta los servicios:
   ```sh
   docker-compose up --build
   ```
2. La API estará disponible en `http://localhost:8080` y la base de datos en el puerto `5432`.

## Notas
- El proyecto realiza migración automática del modelo `User` al iniciar.
- Puedes agregar más modelos y rutas siguiendo la estructura modular. 