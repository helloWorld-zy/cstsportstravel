<template>
  <view class="home">
    <view class="search-bar">
      <input type="text" placeholder="搜索目的地、产品" />
    </view>
    <view class="banner">
      <swiper indicator-dots autoplay circular>
        <swiper-item v-for="i in 3" :key="i">
          <view class="banner-item">Banner {{ i }}</view>
        </swiper-item>
      </swiper>
    </view>
    <view class="section">
      <text class="section-title">推荐产品</text>
    </view>
  </view>
</template>

<script setup lang="ts">
// Homepage
</script>

<style scoped>
.home {
  padding: 16rpx;
}
.search-bar {
  padding: 16rpx;
  background: #fff;
  border-radius: 8rpx;
  margin-bottom: 16rpx;
}
.search-bar input {
  width: 100%;
  font-size: 28rpx;
}
.banner-item {
  height: 300rpx;
  background: #e0e0e0;
  display: flex;
  align-items: center;
  justify-content: center;
  border-radius: 8rpx;
}
.section-title {
  font-size: 32rpx;
  font-weight: bold;
  padding: 16rpx 0;
}
</style>
