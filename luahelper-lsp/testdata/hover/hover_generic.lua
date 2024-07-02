
---@class UnitObject
local UnitClass = class('UnitObject')


---@class HeroObject : UnitObject
local HeroObject = class('HeroObject', UnitObject)


---@class EmptyObject : UnitObject
local EmptyObject = class('EmptyObject', UnitObject)


---@generic T
---@param classType T
---@return T
function GoodsBuild(classType)
end


---@generic T
---@param className `T`
---@return T
function GoodsBuildByName(className)
end


---@generic T
---@param classType T
---@return T[]
function GoodsBuilds(classType)
end


---@generic T
---@param className `T`
---@return T[]
function GoodsBuildsByName(className)
end


local nameLocal   = 'GoodsModule'
local nameGlobal  = GoodsDefine.BaseGoods
local classLocal  = require('hover_generic_class')
local classGlobal = GoodsClass.BaseClass
-- 全局方法定义测试
local cloneGoods1 = clone(EmptyObject)
local cloneGoods2 = clone(classLocal)
local cloneGoods3 = clone(classGlobal)
local cloneGoods4 = clone(GoodsClass.BaseClass)
-- 局部方法定义测试
local classGoods1 = GoodsBuild(classLocal)
local classGoods2 = GoodsBuild(classGlobal)
-- 字符串泛型定义测试
local nameGoods1 = GoodsBuildByName('GoodsModule')
local nameGoods2 = GoodsBuildByName(nameLocal)
local nameGoods3 = GoodsBuildByName(nameGlobal)
local nameGoods4 = GoodsBuildByName(GoodsDefine.BaseGoods)
local nameGoods5 = createByName('GoodsModule')
local nameGoods6 = createByName(nameLocal)
local nameGoods7 = createByName(nameGlobal)
local nameGoods8 = createByName(GoodsDefine.BaseGoods)
-- 数组泛型定义测试
local goodsList1 = GoodsBuilds(classLocal)
local goodsList2 = GoodsBuilds(GoodsClass.BaseClass)
local goodsList3 = GoodsBuildsByName('GoodsModule')
local goodsList4 = GoodsBuildsByName(nameGlobal)
local goodsList5 = GoodsBuildsByName(GoodsDefine.BaseGoods)
local goodsList6 = createListByName('GoodsModule')
local goodsList7 = createListByName(nameGlobal)
local goodsList8 = createListByName(GoodsDefine.BaseGoods)
-- 类成员方法定义测试
local goodsIns11 = Instances.GoodsIns:CloneGoods(EmptyObject)
local goodsIns12 = Instances.GoodsIns.CloneGoods(classLocal)
local goodsIns13 = Instances.GoodsIns:CloneGoods(classGlobal)
local goodsIns14 = Instances.GoodsIns:CloneGoods(GoodsClass.BaseClass)
local goodsIns21 = Instances.GoodsIns:CreateGoods('GoodsModule')
local goodsIns22 = Instances.GoodsIns:CreateGoods(nameLocal)
local goodsIns23 = Instances.GoodsIns:CreateGoods(nameGlobal)
local goodsIns24 = Instances.GoodsIns:CreateGoods(GoodsDefine.BaseGoods)
local goodsIns31 = Instances.GoodsIns:CloneGoodsList(EmptyObject)
local goodsIns32 = Instances.GoodsIns:CloneGoodsList(classLocal)
local goodsIns33 = Instances.GoodsIns:CloneGoodsList(classGlobal)
local goodsIns34 = Instances.GoodsIns:CloneGoodsList(GoodsClass.BaseClass)
local goodsIns41 = Instances.GoodsIns:CreateGoodsList('GoodsModule')
local goodsIns42 = Instances.GoodsIns:CreateGoodsList(nameLocal)
local goodsIns43 = Instances.GoodsIns:CreateGoodsList(nameGlobal)
local goodsIns44 = Instances.GoodsIns:CreateGoodsList(GoodsDefine.BaseGoods)
