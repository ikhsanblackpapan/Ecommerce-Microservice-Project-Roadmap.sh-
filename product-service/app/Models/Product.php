<?php

namespace App\Models;

use Illuminate\Database\Eloquent\Model;
use Illuminate\Support\Str;

class Product extends Model
{
   protected $keyType = 'string'; // Menentukan tipe primary key sebagai string
   public $incrementing = false;

   protected $fillable = ['category_id', 'name', 'description', 'price', 'stock'];

   protected static function booted()
   {
        static::creating(function ($product) {
            $product->id = (string) Str::uuid(); // Menghasilkan UUID saat membuat produk baru
        });
   }

   public function category()
    {
        return $this->belongsTo(Category::class);
    }
}
