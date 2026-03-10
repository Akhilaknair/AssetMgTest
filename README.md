# Asset Management System – (backend API to manage company assets and users)

## Project Features

## 1. User Authentication & Authorization

i) The system provides user registration and login functionality.  
ii) Authenticated users can securely log out of the system.  
iii) Users can view and manage their profile information.  
iv) Users are allowed to delete their own accounts.  
v) Role-based access control is implemented with three roles: Admin, Asset Manager, and Employee.


## 2. User Management

i) Admins and Asset Managers can view the list of all registered users.  
ii) Admins and Asset Managers can update and manage user roles.


## 3. Asset Management

i) Employees can view the assets assigned to them.  
ii) Admins and Asset Managers can create new assets.  
iii) Admins and Asset Managers can update existing asset details.  
iv) Admins and Asset Managers can delete assets from the system.  
v) Assets can be assigned to users by Admins and Asset Managers.  
vi) Assigned assets can be returned and marked accordingly.  
vii) Assets can be marked as “Need Service” when maintenance is required.  
viii) A dashboard view is available to see a summary of all assets.


## 4. Server & Security

i) Protected routes require user authentication.  
ii) Middleware is used to enforce authentication and role-based authorization.  
iii) A health check endpoint is available to verify that the server is running.


## 5. API Routes Available are :

### Health Check

i) `GET /health`  -> Used to verify that the server is running.


### Authentication Routes (`/auth`)-> requires authentication

i) `POST /auth/register`  ->Used to register a new user.

ii) `POST /auth/login` -> Used to authenticate a user and log in.

iii) `POST /auth/logout`  -> Used to log out the currently authenticated user.

iv) `GET /auth/profile`  -> Used to fetch the logged-in user’s profile details.

v) `DELETE /auth/delete`  ->Used to delete the logged-in user account.


### User Management Routes (`/users`)->(Requires authentication and Admin, Asset Manager role)

i) `GET /users/`  -> Used to fetch all users.

ii) `PUT /users/{id}/role`  -> Used to update the role of a specific user.


### Asset Routes (`/assets`)-> (Requires authentication)

i) `GET /assets/`  ->Used to fetch assets assigned to the logged-in employee.

ii) `GET /assets/{id}`  ->Used to fetch details of a specific asset.


###  Asset Management Routes-> (Need authentication & accesed by Admin,Asset Manager Only)

i) `POST /assets/`  ->Used to create a new asset.

ii) `GET /assets/dashboard`  -> Used to fetch asset summary for dashboard view.

iii) `PUT /assets/{id}`  ->Used to update asset details.

iv) `DELETE /assets/{id}`  -> Used to delete an asset.

v) `POST /assets/{id}/assign`  -> Used to assign an asset to a user.

vi) `POST /assets/{id}/return`  -> Used to return an assigned asset.

vii) `PUT /assets/{id}/need-service`  -> Used to mark an asset as needing service.

