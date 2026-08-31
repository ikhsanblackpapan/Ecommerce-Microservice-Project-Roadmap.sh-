<?php

use App\Http\Controllers\Api\ProductController;
use Illuminate\Support\Facades\Route;

// Endpoint otomatis: GET, POST, PUT/PATCH, DELETE untuk /api/products
Route::apiResource('products', ProductController::class);