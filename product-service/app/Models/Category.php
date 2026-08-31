<?php

namespace App\Models;

use Illuminate\Database\Eloquent\Model;

class Category extends Model
{
    // Tambahkan baris ini untuk mengizinkan pengisian kolom 'name'
    protected $fillable = ['name', 'slug']; 

    public function products()
    {
        return $this->hasMany(Product::class);
    }
}
