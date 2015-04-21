!function(namespace) {
  'use strict';

  function Gantt(elem, params) {
    this.$element = jQuery(elem);
    this.params = params || {};

    this.cssHiddenChartLine = this.params.cssHiddenChartLine || 'JS-Gantt-ChartLine-hidden';
    this.cssHiddenChartGuide = this.params.cssHiddenChartGuide || 'JS-Gantt-ChartGuide-hidden';
    this.cssReadyElement = this.params.cssReadyElement || 'JS-Gantt-ready';
    this.timePerPixel = this.params.timePerPixel || null;
    this.timeStep = this.params.timeStep || 24 * 60 * 60 * 1000;

    this.__construct();
  }

  Gantt.prototype.__construct = function __construct() {
    if (typeof Modal === 'undefined') {
      return false;
    }

    this.$viewport = this.$element.find('.JS-Gantt-Viewport');
    this.$chart = this.$viewport.find('.JS-Gantt-Chart');
    this.$chartLineBlank = this.$chart.find('.JS-Gantt-ChartLine').eq(0);
    this.$chartFieldBlank = this.$chartLineBlank.find('.JS-Gantt-ChartField').eq(0);
    this.$chartPartBlank = this.$chartFieldBlank.find('.JS-Gantt-ChartPart').eq(0);
    this.$chartGuideBlank = this.$element.find('.JS-Gantt-ChartGuide').eq(0);
    this.$sidebar = this.$element.find('.JS-Gantt-Sidebar');
    this.$sidebarLineBlank = this.$sidebar.find('.JS-Gantt-SidebarLine');
    this.$modal = jQuery('.JS-Gannt-Modal');

    this.data = {};
    this.startTime = new Date().getTime();
    this.endTime = 0;
    this.pixelDensity = 1000 * 60;
    this.chartLineHeight = this.$chartLineBlank.height();

    this._init();
  };

  Gantt.prototype._init = function _init() {
    var _this = this;

    this.$chart.on('click.JS-Gantt', '.JS-Gantt-ChartField', function() { _this._showDetails.apply(_this, [jQuery(this)]); });

    this._start();
  } ;

  Gantt.prototype._start = function _start() {
    this._getData();
    this._searchStartEndTime();

    if (!this.timePerPixel) {
      this._getTimePerPixel();
    } else {
      this._setChartWidth();
      this._initDraggable();
    }

    this._drawChart();
    this._drawTimeGrid();

    this._ready();
  };

  Gantt.prototype._ready = function _ready() {
    // dev
    // console.log(this.data);

    this.$element
      .addClass(this.cssReadyElement)
      .addClass('JS-Gantt-ready');
  };

  Gantt.prototype._getData = function _getData() {
    var _this = this;

    jQuery.ajax({
      url: '/backups.json',
      type: 'GET',
      async: false,
      success: function(response) {
        _this.data = response;
      },
      error: function() {
        _this.data = {};
      }
    });
  };

  Gantt.prototype._searchStartEndTime = function _searchStartEndTime() {
    var key,
        job,
        task,
        startTime,
        endTime,
        i,
        j;

    for (key in this.data) {
      job = this.data[key];

      for (i = 0, j = job.length; i < j; i++) {
        task = job[i];

        if (typeof task.StartTime !== 'undefined') {
          startTime = new Date(task.StartTime).getTime();

          if ( startTime < this.startTime) {
            this.startTime = startTime;
          }
        }

        if (typeof task.EndTime !== 'undefined') {
          endTime = new Date(task.EndTime).getTime();

          if ( endTime > this.endTime) {
            this.endTime = endTime;
          }
        }
      }
    }
  };

  Gantt.prototype._getTimePerPixel = function _getTimePerPixel() {
    this.timePerPixel = (this.endTime - this.startTime) / this.$chart.width();
  };

  Gantt.prototype._setChartWidth = function _setChartWidth() {
    var width = (this.endTime - this.startTime) / this.timePerPixel;
    this.$chart.css({width: Math.ceil(width) + 'px'});
  };

  Gantt.prototype._initDraggable = function _initDraggable() {
    this.$chart.draggable({
      axis: 'x',
      cursor: 'e-resize',
      scroll: true
    });
  };

  Gantt.prototype._drawChart = function _drawChart() {
    var key,
        job,
        task,
        taskPosition,
        taskCount,
        i,
        j;

    for (key in this.data) {
      job = this.data[key];

      // draw task
      for (i = 0, j = job.length; i < j; i++) {
        task = job[i];

        if (typeof task.StartTime !== 'undefined' && typeof task.EndTime !== 'undefined') {
          taskPosition = this._getTaskPosition(task.StartTime, task.EndTime);
          this._drawTask(taskPosition, task.TaskId, task.JobName);
        }
      }

      // draw job
        this._drawJob(key, j);
    }
  };

  Gantt.prototype._getTaskPosition = function _getTaskPosition(startTime, endTime) {
    if (arguments.length !== 2) {
      return false;
    }

    var taskStart,
        taskEnd,
        startPosition,
        endPosition,
        taskWidth;

    taskStart = new Date(startTime).getTime() - this.startTime;
    taskEnd = new Date(endTime).getTime() - this.startTime;

    startPosition = Math.floor(taskStart / this.timePerPixel);
    endPosition = Math.ceil(taskEnd / this.timePerPixel);
    taskWidth = endPosition - startPosition;

    if (taskWidth < 1) {
      taskWidth = 1;
    }

    return {start: startPosition, width: taskWidth};
  };

  Gantt.prototype._drawTask = function _drawTask(taskPosition, taskId, jobName) {
    if (arguments.length < 3) {
      return false;
    }

    var $line = this.$chartLineBlank.clone(),
        $field = $line.find('.JS-Gantt-ChartField');

    $line
      .removeClass(this.cssHiddenChartLine)
      .removeClass('JS-Gantt-ChartLine-hidden');

    this.$chart.append($line);

    $field
      .css({
        'left': taskPosition.start + 'px',
        'width': taskPosition.width + 'px'
      })
      .attr('data-taskId', taskId)
      .attr('data-jobName', jobName);
  };

  Gantt.prototype._drawJob = function _drawJob(key, taskCount) {
    if (arguments.length !== 2) {
      return false;
    }

    var $line = this.$sidebarLineBlank.clone();

    $line
      .removeClass(this.cssHiddenChartLine)
      .removeClass('JS-Gantt-ChartLine-hidden');

    $line.text(key);

    this.$sidebar.append($line);

    $line.css({
      'height': this.chartLineHeight * taskCount + 'px'
    });
  };

  Gantt.prototype._drawTimeGrid = function _drawTimeGrid() {
    var $guide,
        guidePosition,
        date,
        i;

    for (i = this.startTime; i < this.endTime; i += this.timeStep) {
      $guide = this.$chartGuideBlank.clone();
      guidePosition = (i - this.startTime) / this.timePerPixel;
      date = new Date(i);

      $guide.attr('data-label', date.getDate() + '/' + date.getMonth() + '/' + date.getFullYear());
      $guide
        .removeClass(this.cssHiddenChartGuide)
        .removeClass('JS-Gantt-ChartGuide-hidden');

      this.$chart.append($guide);

      $guide.css({'left': Math.ceil(guidePosition) + 'px'});
    }
  };


  Gantt.prototype._showDetails = function _showDetails($source) {
    if (!arguments.length) {
      return false;
    }

    var taskId = $source.attr('data-taskId'),
        jobName = $source.attr('data-jobName'),
        info,
        i,
        j;

    if (taskId) {
      // search info
      for (i = 0, j = this.data[jobName].length; i < j; i++) {
        if (this.data[jobName][i].TaskId === taskId) {
          info = this.data[jobName][i];
        } else {
          continue;
        }
      }

      // append to popup
      this.$modal.trigger('modal:open');

      // show popup
    }
  };

  namespace.Gantt = Gantt;
}(this);
