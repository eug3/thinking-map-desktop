"use client";
import { useState, useEffect } from "react";
import { Label } from "@/components/ui/label";
import { Textarea } from "@/components/ui/textarea";
import { Button } from "@/components/ui/button";
import { RadioGroup, RadioGroupItem } from "@/components/ui/radio-group";
import { Command, Github, Loader, Clock, Trash2, ArrowRight, Search } from "lucide-react";
import { Dialog, DialogContent, DialogDescription, DialogFooter, DialogHeader, DialogTitle } from "@/components/ui/dialog";
import { AlertDialog, AlertDialogAction, AlertDialogCancel, AlertDialogContent, AlertDialogDescription, AlertDialogFooter, AlertDialogHeader, AlertDialogTitle, AlertDialogTrigger } from "@/components/ui/alert-dialog";
import { Card, CardContent, CardDescription, CardFooter, CardHeader, CardTitle } from "@/components/ui/card";
import { Badge } from "@/components/ui/badge";
import { Input } from "@/components/ui/input";
import { toast } from "sonner";
import { useNavigate } from "react-router-dom";
import { createMap, fetchMapList, deleteMap } from "@/api";
import type { CreateMapRequest, Map as ThinkingMap, MapListQuery } from "@/types";
import SettingsDialog from "@/components/SettingsDialog";

const PROBLEM_TYPES = [
  { value: '研究型', label: '研究型', description: '深入探索和分析特定主题' },
  { value: '创意型', label: '创意型', description: '发散思维，寻找创新解决方案' },
  { value: '分析型', label: '分析型', description: '系统分析数据和现象' },
  { value: '规划型', label: '规划型', description: '制定策略和执行计划' }
];

const PROBLEM_EXAMPLES = {
  研究型: '评估人工智能在教育领域的应用效果？',
  创意型: '设计一个创新的城市共享单车系统',
  分析型: '分析全球供应链中断对电子产品市场的影响',
  规划型: 'python入门学习规划'
};
const GITHUB_URL = "https://github.com/PGshen/thinking-map";

export default function HomePage() {
  const [problem, setProblem] = useState('');
  const [problemType, setProblemType] = useState('研究型');

  const [isDialogOpen, setIsDialogOpen] = useState(false);
  const [title, setTitle] = useState('');
  const [goal, setGoal] = useState('');
  const [isLoading, _setIsLoading] = useState(false);

  // 历史列表状态
  const [mapList, setMapList] = useState<ThinkingMap[]>([]);
  const [listLoading, setListLoading] = useState(false);
  const [searchText, setSearchText] = useState('');
  const [totalMaps, setTotalMaps] = useState(0);
  const [currentPage, setCurrentPage] = useState(1);
  const pageSize = 12;

  const navigate = useNavigate();

  // 加载工作区列表
  const loadMapList = async (page = 1, search = '') => {
    setListLoading(true);
    try {
      const query: MapListQuery = {
        page,
        limit: pageSize,
        search: search || undefined,
      };
      const result = await fetchMapList(query);
      setMapList(result.items || []);
      setTotalMaps(result.total || 0);
      setCurrentPage(page);
    } catch (e) {
      console.error('加载工作区列表失败', e);
    } finally {
      setListLoading(false);
    }
  };

  // 初始加载
  useEffect(() => {
    loadMapList();
  }, []);

  // 删除工作区
  const handleDeleteMap = async (mapId: string) => {
    try {
      await deleteMap(mapId);
      toast.success('删除成功');
      loadMapList(currentPage, searchText);
    } catch (e) {
      toast.error('删除失败');
    }
  };

  // 搜索
  const handleSearch = () => {
    loadMapList(1, searchText);
  };

  const handleCreateMap = async () => {
    if (!problem.trim()) {
      toast.error("问题描述不能为空");
      return;
    }
    const params: CreateMapRequest = {
      title: title.trim() || problem.trim().substring(0, 50),
      problem: problem.trim(),
      problemType: problemType || undefined,
      target: goal.trim() || undefined,
      keyPoints: undefined,
      constraints: undefined,
    };
    try {
      const res = await createMap(params);
      if (res && res.id) {
        setIsDialogOpen(false);
        loadMapList(); // 刷新列表
        navigate(`/workspace/${res.id}`);
      } else {
        toast.error("创建思维导图失败");
      }
    } catch (e) {
      toast.error("创建思维导图失败");
    }
  };

  const handleSubmit = () => {
    if (problem.length < 6) {
      toast.info("问题描述至少需要6个字，请补充更多细节。");
      return;
    }
    setIsDialogOpen(true);
  };

  return (
    <div className="container mx-auto px-4 py-8 flex flex-col items-center min-h-screen">
      <SettingsDialog />
      <a
        href={GITHUB_URL}
        target="_blank"
        rel="noopener noreferrer"
        aria-label="打开 GitHub 仓库"
        className="fixed top-3 right-3 inline-flex items-center gap-1 rounded-full bg-black text-white px-2 py-2 shadow hover:bg-gray-900"
      >
        <Github className="h-4 w-4" />
        <span className="text-xs font-medium">GitHub</span>
      </a>

      {/* 品牌展示区 */}
      <div className="text-center mb-8">
        <div className="flex mb-2 items-end justify-center">
          <div className="mr-4 flex size-10 items-center justify-center rounded-xl bg-primary text-primary-foreground shadow-lg">
            <Command className="size-6" />
          </div>
          <h1 className="text-4xl font-bold">ThinkingMap Desktop</h1>
        </div>
        <p className="text-lg text-gray-600">可视化解构您的思维历程，让问题解决变得透明和可控</p>
      </div>

      {/* 问题输入区 */}
      <div className="w-full max-w-2xl space-y-3">
        <div className="relative">
          <Textarea
            placeholder="请输入您的问题（50-200字最佳）..."
            value={problem}
            onChange={(e) => setProblem(e.target.value)}
            className="min-h-[120px] pr-24"
          />
          <div className="absolute bottom-2 right-2 flex items-center space-x-2">
            <span className="text-sm text-gray-500">
              {problem.length} / 200 字
              {problem.length > 200 && (
                <span className="text-yellow-500 ml-2">字数超出建议范围</span>
              )}
            </span>
            <Button
              onClick={handleSubmit}
              disabled={!problem.trim() || !problemType}
              className="ml-2 cursor-pointer"
              size="sm"
            >
              开始分析
            </Button>
          </div>
        </div>

        <div className="space-y-3">
          <Label className="text-base">选择问题类型</Label>
          <RadioGroup
            value={problemType}
            onValueChange={setProblemType}
            className="grid grid-cols-1 md:grid-cols-2 gap-1"
          >
            {PROBLEM_TYPES.map((type) => (
              <div key={type.value} className="flex items-center space-x-2 p-1 rounded-lg hover:bg-slate-50">
                <RadioGroupItem value={type.value} id={type.value} />
                <Label htmlFor={type.value} className="flex-1 cursor-pointer">
                  <span className="font-medium">{type.label}</span>
                  <span className="text-gray-500 ml-2">- {type.description}</span>
                </Label>
              </div>
            ))}
          </RadioGroup>
        </div>

        {problemType && (
          <div className="text-sm text-gray-600">
            <p>示例问题：{PROBLEM_EXAMPLES[problemType as keyof typeof PROBLEM_EXAMPLES]}</p>
          </div>
        )}
      </div>

      {/* 历史工作区列表 */}
      <div className="w-full max-w-4xl mt-12">
        <div className="flex items-center justify-between mb-4">
          <div className="flex items-center gap-2">
            <Clock className="h-5 w-5 text-gray-500" />
            <h2 className="text-xl font-semibold">历史工作区</h2>
            {totalMaps > 0 && (
              <span className="text-sm text-gray-400">共 {totalMaps} 个</span>
            )}
          </div>
          <div className="flex items-center gap-2">
            <Input
              placeholder="搜索问题..."
              value={searchText}
              onChange={(e) => setSearchText(e.target.value)}
              onKeyDown={(e) => e.key === 'Enter' && handleSearch()}
              className="w-48 h-8 text-sm"
            />
            <Button variant="outline" size="sm" onClick={handleSearch} className="h-8 cursor-pointer">
              <Search className="h-3 w-3" />
            </Button>
          </div>
        </div>

        {listLoading ? (
          <div className="flex justify-center py-12">
            <Loader className="h-6 w-6 animate-spin text-gray-400" />
          </div>
        ) : mapList.length === 0 ? (
          <div className="text-center py-12 text-gray-400">
            <p>暂无历史工作区，创建一个新的开始吧！</p>
          </div>
        ) : (
          <>
            <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-4">
              {mapList.map((map) => (
                <Card
                  key={map.id}
                  className="hover:shadow-md transition-shadow cursor-pointer group"
                  onClick={() => navigate(`/workspace/${map.id}`)}
                >
                  <CardHeader className="pb-2">
                    <div className="flex items-start justify-between">
                      <CardTitle className="text-base line-clamp-1">{map.title || '未命名'}</CardTitle>
                      <StatusBadge status={map.status} />
                    </div>
                    {map.problemType && (
                      <CardDescription className="text-xs">{map.problemType}</CardDescription>
                    )}
                  </CardHeader>
                  <CardContent className="pb-2">
                    <p className="text-sm text-gray-600 line-clamp-2">{map.problem}</p>
                  </CardContent>
                  <CardFooter className="pt-0 flex items-center justify-between">
                    <span className="text-xs text-gray-400">
                      {formatTime(map.createdAt)}
                    </span>
                    <div className="flex items-center gap-1 opacity-0 group-hover:opacity-100 transition-opacity">
                      <AlertDialog>
                        <AlertDialogTrigger asChild>
                          <Button
                            variant="ghost"
                            size="icon"
                            className="h-7 w-7 cursor-pointer"
                            onClick={(e) => e.stopPropagation()}
                          >
                            <Trash2 className="h-3.5 w-3.5 text-red-400" />
                          </Button>
                        </AlertDialogTrigger>
                        <AlertDialogContent onClick={(e) => e.stopPropagation()}>
                          <AlertDialogHeader>
                            <AlertDialogTitle>确认删除</AlertDialogTitle>
                            <AlertDialogDescription>
                              确定要删除「{map.title || '未命名'}」吗？此操作不可撤销。
                            </AlertDialogDescription>
                          </AlertDialogHeader>
                          <AlertDialogFooter>
                            <AlertDialogCancel>取消</AlertDialogCancel>
                            <AlertDialogAction onClick={() => handleDeleteMap(map.id)}>
                              删除
                            </AlertDialogAction>
                          </AlertDialogFooter>
                        </AlertDialogContent>
                      </AlertDialog>
                      <Button variant="ghost" size="icon" className="h-7 w-7 cursor-pointer">
                        <ArrowRight className="h-3.5 w-3.5" />
                      </Button>
                    </div>
                  </CardFooter>
                </Card>
              ))}
            </div>

            {/* 分页 */}
            {totalMaps > pageSize && (
              <div className="flex justify-center gap-2 mt-6">
                <Button
                  variant="outline"
                  size="sm"
                  disabled={currentPage <= 1}
                  onClick={() => loadMapList(currentPage - 1, searchText)}
                  className="cursor-pointer"
                >
                  上一页
                </Button>
                <span className="text-sm text-gray-500 flex items-center px-2">
                  {currentPage} / {Math.ceil(totalMaps / pageSize)}
                </span>
                <Button
                  variant="outline"
                  size="sm"
                  disabled={currentPage >= Math.ceil(totalMaps / pageSize)}
                  onClick={() => loadMapList(currentPage + 1, searchText)}
                  className="cursor-pointer"
                >
                  下一页
                </Button>
              </div>
            )}
          </>
        )}
      </div>

      {/* 问题理解确认对话框 */}
      <Dialog open={isDialogOpen} onOpenChange={setIsDialogOpen}>
        <DialogContent className="max-w-2xl">
          <DialogHeader>
            <DialogTitle>创建思维导图</DialogTitle>
            <DialogDescription>确认问题并开始分析</DialogDescription>
          </DialogHeader>

          <div className="space-y-4 py-4">
            <div className="space-y-2">
              <Label>标题</Label>
              <Textarea
                value={title}
                onChange={(e) => setTitle(e.target.value)}
                placeholder="思维导图标题（可选）"
                className="min-h-[60px]"
              />
            </div>
            <div className="space-y-2">
              <Label>问题描述</Label>
              <Textarea
                value={problem}
                onChange={(e) => setProblem(e.target.value)}
                className="min-h-[80px]"
              />
            </div>

            <div className="space-y-2">
              <Label>目标（可选）</Label>
              <Textarea
                value={goal}
                onChange={(e) => setGoal(e.target.value)}
                placeholder="您希望达到什么目标？"
                className="min-h-[60px]"
              />
            </div>
          </div>

          <DialogFooter className="flex-row justify-end space-x-2">
            <Button variant="outline" onClick={() => setIsDialogOpen(false)}>
              取消
            </Button>
            <Button onClick={handleCreateMap} disabled={isLoading}>
              {isLoading && <Loader className="h-4 w-4 mr-2 animate-spin" />}
              创建思维导图
            </Button>
          </DialogFooter>
        </DialogContent>
      </Dialog>
    </div>
  );
}

// 状态标签组件
function StatusBadge({ status }: { status: string }) {
  const map: Record<string, { label: string; variant: 'default' | 'secondary' | 'destructive' | 'outline' }> = {
    initial: { label: '待开始', variant: 'outline' },
    running: { label: '进行中', variant: 'default' },
    completed: { label: '已完成', variant: 'secondary' },
    in_decomposition: { label: '拆解中', variant: 'default' },
    in_conclusion: { label: '总结中', variant: 'default' },
  };
  const info = map[status] || { label: status, variant: 'outline' as const };
  return <Badge variant={info.variant} className="text-[10px] h-5">{info.label}</Badge>;
}

// 格式化时间
function formatTime(timeStr: string | number) {
  if (!timeStr) return '';
  let date: Date;
  if (typeof timeStr === 'number') {
    // Unix timestamp (seconds or milliseconds)
    date = new Date(timeStr > 1e12 ? timeStr : timeStr * 1000);
  } else {
    date = new Date(timeStr);
  }
  if (isNaN(date.getTime())) return String(timeStr);
  const now = new Date();
  const diffMs = now.getTime() - date.getTime();
  const diffMin = Math.floor(diffMs / 60000);
  if (diffMin < 1) return '刚刚';
  if (diffMin < 60) return `${diffMin}分钟前`;
  const diffHour = Math.floor(diffMin / 60);
  if (diffHour < 24) return `${diffHour}小时前`;
  const diffDay = Math.floor(diffHour / 24);
  if (diffDay < 30) return `${diffDay}天前`;
  return date.toLocaleDateString('zh-CN');
}
